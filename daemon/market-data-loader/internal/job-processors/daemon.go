package jobprocessors

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gofr.dev/pkg/gofr"

	dataProviders "github.com/stratifyr/security-service/daemon/stats-loader/internal/data-providers"
)

type daemon struct {
	dataProvider dataProviders.Provider
}

func NewDaemon(dataProvider dataProviders.Provider) *daemon {
	return &daemon{dataProvider: dataProvider}
}

func (d *daemon) Start(ctx *gofr.Context) error {
	go func() {
		c, stop := signal.NotifyContext(ctx.Context, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				ctx.Logger.Debug("looking for pending jobs...")

				if err := d.processMarketDataJob(ctx); err != nil {
					ctx.Logger.Warnf("failed to process market data job: %v", err)
				}
			case <-c.Done():
				ctx.Logger.Info("shutting down daemon...")
				return
			}
		}
	}()

	return nil
}

func (d *daemon) processMarketDataJob(ctx *gofr.Context) error {
	job, err := GetNextPendingJob(ctx)
	if err != nil {
		return err
	}

	if job == nil {
		ctx.Logger.Debug("no pending job found")
		return nil
	}

	ctx.Logger.Infof("processing %s job id:%d", strings.ToLower(job.Type), job.ID)

	jobProcessor, err := GetJobProcessor(job.Type, d.dataProvider)
	if err != nil {
		return err
	}

	logs, err := jobProcessor.Process(ctx)
	if err != nil {
		_ = UpdateJobStatus(ctx, job.ID, "FAILED", logs)
		return err
	}

	_ = UpdateJobStatus(ctx, job.ID, "COMPLETED", logs)

	ctx.Logger.Infof("processed %s job id:%d", strings.ToLower(job.Type), job.ID)

	return nil
}
