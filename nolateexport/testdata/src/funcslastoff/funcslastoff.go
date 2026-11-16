package funcslastoff

// Same shape as the funcslast package, analyzed with the zero Config:
// unexported-funcs-last is off by default, so only same-type ordering is checked.

type Runner struct{}

func stableGitEnv() []string { return nil } // OK

func (r *Runner) Run() {} // OK

func (r *Runner) Close() {} // OK

type Other struct{}

func (o Other) helper() {} // OK

func (o Other) Late() {} // want `exported method Other.Late should appear before unexported methods of Other`
