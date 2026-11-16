package funcslast

type Runner struct{}

func stableGitEnv() []string { return nil } // OK

func (r *Runner) Run() {} // want `exported method Runner.Run should appear before unexported standalone functions`

func (r *Runner) Close() {} // want `exported method Runner.Close should appear before unexported standalone functions`

// Each method gets at most one diagnostic: the per-type message wins.

type Other struct{}

func (o Other) helper() {} // OK

func (o Other) Late() {} // want `exported method Other.Late should appear before unexported methods of Other`
