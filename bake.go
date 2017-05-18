package main

import (
	"fmt"
	"github.com/Sirupsen/logrus"
	"gopkg.in/cheggaaa/pb.v1"
	"sync"
)

func Bake(workers int) error {
	tasks := make(chan string)
	queue := make(chan string)
	finished := make(chan string)
	finishedCount := 0

	log.Debug(fmt.Sprintf("Using %d workers", workers))

	files, dependencies, err := NeedingCompilation(repository)
	// BADBAD: need to display error from compilation if it fails
	if err != nil {
		return err
	}

	var bar *pb.ProgressBar
	if log.Level != logrus.DebugLevel {
		bar = pb.StartNew(len(files))
		//bar.SetMaxWidth(80)
		bar.ShowTimeLeft = true
		bar.Start()
	}

	// Manage a pool of workers
	var group sync.WaitGroup
	for i := 0; i < workers; i++ {
		group.Add(1)
		go func() {
			for task := range tasks {
				log.Debug(fmt.Sprintf("Worker %d is compiling %s", i, task))
				Compile(repository, task)

				// Non-blocking send
				finishedCount++
				select {
				case finished <- task:
				default:
				}

				if log.Level != logrus.DebugLevel {
					bar.Increment()
				} else {
					log.Info(fmt.Sprintf("Finished %d/%d tasks", finishedCount, len(files)))
				}
			}
			group.Done()
		}()
	}

	// As things arrive in the queue, send them out to the various workers
	go func() {
		for file := range queue {
			log.Debug("Sending " + file + " to workers")
			tasks <- file
		}
		close(tasks)
	}()

	compiled := make(map[string]bool)
	enqueued := make(map[string]bool)
	enqueuedCount := 0

	for {
		// add everything that can be compiled given what has been completed
		for _, file := range files {
			if enqueued[file] {
				continue
			}

			good := true
			for _, dependency := range dependencies[file] {
				if !compiled[dependency] {
					good = false
					break
				}
			}

			if good {
				log.Debug("Placing " + file + " into the queue")
				queue <- file
				enqueued[file] = true
				enqueuedCount++
			}
		}

		log.Debug(fmt.Sprintf("%d tasks have been placed in the queue.", enqueuedCount))
		if enqueuedCount >= len(files) {
			log.Debug("Everything is in the queue.")
			break
		}

		log.Debug("Wait for things to finish before placing more into the queue.")
		// get more finished things
		compiled[<-finished] = true
	}

	close(queue)
	log.Debug("Waiting for the workers to finish.")
	group.Wait()

	if log.Level != logrus.DebugLevel {
		bar.FinishPrint("The xake is made.")
	} else {
		log.Info("The xake is made.")
	}

	return nil
}
