package parse

import (
	"strconv"
	"strings"
	"time"

	"github.com/Abhra303/quickmap/pkg/datastore"
	"github.com/efficientgo/core/errors"
)

type ExprType string
type SetCondition string

const (
	ExprTypeSet   ExprType = "SET"
	ExprTypeGet   ExprType = "GET"
	ExprTypeQPush ExprType = "QPUSH"
	ExprTypeQPop  ExprType = "QPOP"
	ExprTypeBQPop ExprType = "BQPOP"
)

const (
	SetConditionXX SetCondition = "XX"
	SetConditionNX SetCondition = "NX"
)

type Expr interface {
	Type() ExprType
	Usage() string

	// a specific method so that Expr doesn't accidently
	// implement other interfaces having Type() and
	// Usage() methods.
	QMapExpr()
}

type SetExpr struct {
	Key       interface{}
	Value     interface{}
	Expiry    time.Duration
	Condition datastore.SetCondition
}

type GetExpr struct {
	Key interface{}
}

type QPushExpr struct {
	Key    interface{}
	Values []interface{}
}

type QPopExpr struct {
	Key interface{}
}

type BQPopExpr struct {
	Key     interface{}
	Timeout time.Duration
}

func (e *GetExpr) Type() ExprType   { return ExprTypeGet }
func (e *SetExpr) Type() ExprType   { return ExprTypeSet }
func (e *QPushExpr) Type() ExprType { return ExprTypeQPush }
func (e *QPopExpr) Type() ExprType  { return ExprTypeQPop }
func (e *BQPopExpr) Type() ExprType { return ExprTypeBQPop }

func (e *GetExpr) QMapExpr()   {}
func (e *SetExpr) QMapExpr()   {}
func (e *QPushExpr) QMapExpr() {}
func (e *QPopExpr) QMapExpr()  {}
func (e *BQPopExpr) QMapExpr() {}

func (e *GetExpr) Usage() string {
	return `Usage: GET <key>
	
	Get the value associated with the specified key.
	
	The supported types for <key> are string, float, int, and
	complex numbers.`
}

func (e *SetExpr) Usage() string {
	return `Usage: SET <key> <value> [EX <expiry>] [<condition>]
	
	Set the Key Value pair with optional expiry duration and condition.
	
	The supported types for <key> and <value> are string, float, int,
	and complex numbers.

	<expiry> is an optional integer value which denotes the duration in
	second after which the key-value pair in the store will expire.
	
	<condition> is an optional field used to control the SET behaviour.
	Two supported values are NX and XX.
	NX - Only set the key if it does not already exist
	XX - Only set the key if it already exists`
}

func (e *QPushExpr) Usage() string {
	return `Usage: QPUSH <key> <value...>
	
	Creates a queue if not already created with the <key> and
	appends values to it.
	
	<key> - Name of the queue to write to.
	<value...> - Variadic input that receives multiple values separated by space.`
}

func (e *QPopExpr) Usage() string {
	return `Usage: QPOP <key>

	Returns the last inserted value from the queue with the <key>.
	`
}

func (e *BQPopExpr) Usage() string {
	return `Usage: BQPOP <key> <timeout>
	
	[Experimental] Blocking queue read operation that blocks the thread
	until a value is read from the queue.

	The command will fail if multiple clients try to read from the queue
	at the same time.
	
	<key> - Name of the queue to read from.
	<timeout> - The duration in seconds to wait until a value is read from
	the queue. The argument must be interpreted as a double value. A value
	of 0 immediately returns a value from the queue without blocking.`
}

var InvalidCommandError = errors.New("invalid command")
var UnknownCommandError = errors.New("unknown command")
var UnsupportedCommandError = errors.New("unsupported command")

func detectLiteralType(word string) interface{} {
	i, err := strconv.ParseInt(word, 0, 64)
	if err == nil {
		return i
	}
	j, err := strconv.ParseFloat(word, 64)
	if err == nil {
		return j
	}
	k, err := strconv.ParseComplex(word, 64)
	if err == nil {
		return k
	}
	return word
}

func ParseCommand(cmd string) (Expr, error) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return nil, errors.Wrap(InvalidCommandError, "command can't be empty")
	}
	words := strings.Split(cmd, " ")
	l := len(words)
	switch ExprType(words[0]) {
	case ExprTypeGet:
		if l != 2 {
			return nil, errors.Wrap(InvalidCommandError, "GET should have exactly one argument")
		}
		return &GetExpr{Key: detectLiteralType(words[1])}, nil
	case ExprTypeSet:
		if l < 3 && l > 6 {
			return nil, errors.Wrap(InvalidCommandError, "arguments don't match with SET command usage")
		}
		key := detectLiteralType(words[1])
		value := detectLiteralType(words[2])
		var condition datastore.SetCondition
		var expiry float64

		if l == 4 {
			switch SetCondition(words[3]) {
			case SetConditionXX:
				condition = datastore.SetIfExists
			case SetConditionNX:
				condition = datastore.SetIfNotExists
			default:
				return nil, errors.Wrapf(InvalidCommandError, "unknown condition \"%s\"", words[3])
			}
		}
		if l == 5 || l == 6 {
			if words[3] != "EX" {
				return nil, errors.Wrap(InvalidCommandError, "unknown SET arguments")
			}
			f, err := strconv.ParseFloat(words[4], 64)
			if err != nil {
				return nil, errors.Wrapf(InvalidCommandError, "non float expiry time \"%s\"", words[4])
			}
			expiry = f
			if l == 6 {
				switch SetCondition(words[5]) {
				case SetConditionXX:
					condition = datastore.SetIfExists
				case SetConditionNX:
					condition = datastore.SetIfNotExists
				default:
					return nil, errors.Wrapf(InvalidCommandError, "unknown condition \"%s\"", words[5])
				}
			}
		}
		return &SetExpr{Key: key, Value: value, Expiry: time.Duration(expiry), Condition: condition}, nil
	case ExprTypeQPush:
		if l < 3 {
			return nil, errors.Wrap(InvalidCommandError, "insufficient arguments for QPUSH")
		}
		key := detectLiteralType(words[1])
		values := make([]interface{}, 0, 1)
		for i := 2; i < l; i++ {
			value := detectLiteralType(words[i])
			values = append(values, value)
		}
		return &QPushExpr{Key: key, Values: values}, nil
	case ExprTypeQPop:
		if l != 2 {
			return nil, errors.Wrap(InvalidCommandError, "QPOP should have exactly two arguments")
		}
		key := detectLiteralType(words[1])
		return &QPopExpr{Key: key}, nil
	case ExprTypeBQPop:
		if l != 3 {
			return nil, errors.Wrap(InvalidCommandError, "BQPOP should have exactly three arguments")
		}
		key := detectLiteralType(words[1])
		d, err := strconv.ParseFloat(words[2], 64)
		if err != nil {
			return nil, errors.Wrapf(InvalidCommandError, "non float expiry time \"%s\"", words[2])
		}
		return &BQPopExpr{Key: key, Timeout: time.Duration(d)}, nil
	default:
		return nil, errors.Wrap(UnknownCommandError, words[0])
	}
}
