package jobs

import "github.com/HorseArcher567/octopus/pkg/assemble"

func Register(target, apiURL string) assemble.Domain {
	return func(ctx *assemble.DomainContext) error {
		if err := registerRPCJobs(ctx, target); err != nil {
			return err
		}
		return registerHTTPJobs(ctx, apiURL)
	}
}
