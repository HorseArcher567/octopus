# pkg/app

`pkg/app` is the orchestration kernel of Octopus.

It is responsible for:

- module graph ordering
- phased module execution (`Build`, `RegisterRPC`, `RegisterHTTP`, `RegisterJobs`, `Run`)
- startup and shutdown hooks
- wiring phase-specific capabilities over injected runtimes

It is intentionally not responsible for constructing concrete RPC / HTTP / resource implementations.
