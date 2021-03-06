// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License.

package worker

import (
	"context"
	"sync"

	"github.com/Azure/acr-cli/cmd/api"
)

// WorkerQueue represents the queue of the workers
var WorkerQueue chan chan PurgeJob
var workers []PurgeWorker

// StartDispatcher creates the workers and a goroutine to continuously fetch jobs for them.
func StartDispatcher(ctx context.Context, wg *sync.WaitGroup, acrClient api.AcrCLIClientInterface, nWorkers int) {
	WorkerQueue = make(chan chan PurgeJob, nWorkers)
	workers = []PurgeWorker{}
	for i := 0; i < nWorkers; i++ {
		worker := NewPurgeWorker(wg, WorkerQueue, acrClient)
		worker.Start(ctx)
		workers = append(workers, worker)
	}

	go func() {
		for job := range JobQueue {
			// Get a job from the JobQueue
			go func(job PurgeJob) {
				// Get a worker (block if there are none available) to process the job
				worker := <-WorkerQueue
				// Assign the job to the worker
				worker <- job
			}(job)
		}
	}()
}

// StopDispatcher stops all the workers.
func StopDispatcher() {
	for _, worker := range workers {
		worker.Stop()
	}
}
