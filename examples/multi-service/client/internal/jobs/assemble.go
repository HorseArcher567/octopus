package jobs

import "github.com/HorseArcher567/octopus/pkg/assemble"

func Assemble(target, apiURL string) assemble.Action {
	return func(ctx *assemble.Context) error {
		if err := registerRPCJobs(ctx, target); err != nil {
			return err
		}
		return registerHTTPJobs(ctx, apiURL)
	}
}
