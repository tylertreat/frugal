package common

import (
	"fmt"
	"reflect"

	"github.com/Workiva/frugal/lib/go"
	"github.com/nats-io/go-nats"
)

const (
	preambleHeader = "preamble"
	rambleHeader   = "ramble"
)

func getNatsConn() *nats.Conn {
	addr := nats.DefaultURL
	natsOptions := nats.DefaultOptions
	natsOptions.Servers = []string{addr}
	natsOptions.Secure = false
	conn, err := natsOptions.Connect()
	if err != nil {
		panic(err)
	}

	return conn
}

func clientLoggingMiddleware(called chan<- bool) frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args frugal.Arguments) frugal.Results {
			var (
				isByteArray bool
				length      int
				ret         frugal.Results
			)

			select {
			case called <- true:
			default:
			}
			// If first argument after the fContext is a byte array
			// print the length (size) instead of the contents
			if len(args) > 1 {
				byteArray, ok := args[1].([]byte)
				if ok {
					isByteArray = true
					length = len(byteArray)
				}
			}
			if isByteArray {
				fmt.Printf("%v(%v) = ", method.Name, length)
				ret = next(service, method, args)
				byteArray, ok := ret[0].([]byte)
				if ok {
					length = len(byteArray)
				}
				fmt.Printf("%v\n", length)
			} else {
				fmt.Printf("%v(%v) = ", method.Name, args[1:])
				ret = next(service, method, args)
				fmt.Printf("%v\n", ret[:1])
			}
			return ret
		}
	}
}

func serverLoggingMiddleware(called chan<- bool) frugal.ServiceMiddleware {
	return func(next frugal.InvocationHandler) frugal.InvocationHandler {
		return func(service reflect.Value, method reflect.Method, args frugal.Arguments) frugal.Results {
			var (
				isByteArray bool
				length      int
				ret         frugal.Results
			)

			select {
			case called <- true:
			default:
			}

			ret = next(service, method, args)

			if len(args) > 1 {
				byteArray, ok := args[1].([]byte)
				if ok {
					isByteArray = true
					length = len(byteArray)
				}
			}

			if isByteArray {
				byteArray, ok := ret[0].([]byte)
				if ok {
					isByteArray = true
					retLength := len(byteArray)
					fmt.Printf("%v(%v) = ", method.Name, length)
					fmt.Printf("%v\n", retLength)
				}
			} else {
				fmt.Printf("%v(%v) = ", method.Name, args[1:])
				fmt.Printf("%v\n", ret[:1])
			}
			return ret
		}
	}
}
