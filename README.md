# golang_snippets_learning
Examples of fundamental concepts, syntax, functions, features, etc.

golang_builder.sh - used to build go files for several architectures and operating systems

## Examples 1-4

Introduce basic syntax, basic fmt, os, log package usage 

## Example 5

Add ioutil . more log, and bufio, including: ioutil.ReadFile() , ioutil.WriteFile() , log.Fatalf , log.Panicf (and log.Panicln() ) 
Also bufio.NewReader(os.Stdin) , defer keyword, inputReader.ReadString()

## Logger


Shows how to setup and use a logger using logrus, with a custom formatter for INFO/WARN/DEBUG/ and custom timestamp. 

## Safe Concurrency

Here I try to demonstrate methods of concurrency that (hopefully) should avoid race conditions / deadlocks / resource leaking . Ths might come at the cost of some CPU overhead. We will use the same logger / formatter used above. Please note that for this section I won't be going over mutexes or waitgroups here. They are fundamental concurrency concepts, but the aim of this section is to be safe, not to be comprehensive. I will demonstrate them later. I'll also give more information on patterns at a later point.  Here  I demonstrate:

### No concurrency
 
See noConcurrency.go - Just copying one directory to another. No concurrency. Baseline.

### Channels with Goroutines

see channels_with_goroutines/exampleChannelsWithGoroutines.go to build off of the noConcurrency.go example.

see channels_with_goroutines/imageFetcher/ directory for a more concrete example.

Goroutines and Channels by themselves are used as examples. Channels with goroutines are used together in safety because (to my understanding):

1) No explicit locking - avoiding the use of mutexes (which can cause deadlocks / race conditions)
2) Synchronization - Channels sync access to shared data, so only a single goroutine processes data at a time.
3) Blocking - Send/receive operations to/from a channel blocks a goroutine UNTIL the process can proceed. If done correctly, this should prevent race conditions.




### TODO Errgroups with Contexts

see ctxsWithErrgroups/batchAndProcessXML.go

In short, contexts make it easy to manage/propgate cancellations/timeouts across goroutines. Propogation is sending cancellation signals / timeouts / deadlines to goroutines derived from it. Errgroups make it easy to launch goroutines and capture errors produced by them. Combining the two allows a group of goroutines to be cancelled in the case of a fail/timeout/deadline . 

### TODO Worker Pools

see safe_concurrency/exampleWorkerPool.go

Not difficult to understand - you can limit the number of concurrent. In my opinion, because of encapsulation, it is easier to write/understand/manage. 
