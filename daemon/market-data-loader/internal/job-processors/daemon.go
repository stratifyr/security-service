package jobprocessors

import (
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"gofr.dev/pkg/gofr"

	"github.com/stratifyr/security-service/client"

	dataProviders "github.com/stratifyr/security-service/daemon/stats-loader/internal/data-providers"
)

type daemon struct {
	dataProvider          dataProviders.Provider
	securityServiceClient client.SecurityServiceClient
}

func NewDaemon(dataProvider dataProviders.Provider, securityServiceClient client.SecurityServiceClient) *daemon {
	return &daemon{dataProvider: dataProvider, securityServiceClient: securityServiceClient}
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
	pendingJobs, err := d.securityServiceClient.GetMarketDataJobs(ctx, "CREATED")
	if err != nil {
		return err
	}

	if len(pendingJobs) < 1 {
		ctx.Logger.Debug("no pending job found")
		return nil
	}

	job := pendingJobs[0]

	ctx.Logger.Infof("processing %s job id:%d", strings.ToLower(job.Type), job.Id)

	jobProcessor, err := GetJobProcessor(job.Type, d.dataProvider, d.securityServiceClient)
	if err != nil {
		return err
	}

	logs, err := jobProcessor.Process(ctx)
	if err != nil {
		_ = d.securityServiceClient.UpdateMarketDataJobStatus(ctx, job.Id, "FAILED", logs)
		return err
	}

	_ = d.securityServiceClient.UpdateMarketDataJobStatus(ctx, job.Id, "COMPLETED", logs)

	ctx.Logger.Infof("processed %s job id:%d", strings.ToLower(job.Type), job.Id)

	return nil
}
