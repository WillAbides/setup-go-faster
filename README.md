# Setup Go Faster

It's like actions/setup-go but faster.

### Faster

On Ubuntu runners, setup-go-faster typically takes about 4s vs 10s for setup-go. This is difficult to benchmark on MacOS
and Windows because the action runners are inconsistent from one run to another.

When using a pre-installed version of go, the times are in the 1s range for both setup-go and setup-go-faster.

The performance improvement is achieved by using simple bash scripts instead of nodejs meaning there is less overhead
to deal with.

The exception to the bash-only rule is setup-go-faster downloads and runs https://github.com/WillAbides/semver-select
to evaluate some version constraints. This takes about 700-1000ms and only affects workflows that use semver constraints.

### Install tip

Setup-go-faster will install go tip from source if you set `go-version: tip`.

### New versions available immediately

No need to wait around for another repo to merge a PR when a new version of go is released. Setup-go-faster gets
available versions directly from https://golang.org/dl. As soon as a release is available there, it\'s available to your
workflow.

### Check out the outputs

Look at those outputs. If you want to use GOPATH or GOMODCACHE as input in some other step, you can just grab it from
setup-go-faster\'s output instead of having to add another step just to set an environment variable.

### What\'s missing?

Just the `stable` input. I don\'t understand what `stable` adds for actions/setup-go. If you only want stable builds
you can set go-version accordingly. If there is good use case for `stable`, it can be added.

## Inputs

### go-version

__Required__

The version of go to install. It can be an exact version or a semver constraint like '1.14.x' or '^1.14.4'.
Do not add "go" or "v" to the beginning of the version.

Action runners come with some versions of go pre-installed. If any of those versions meet your semver constraint
setup-go-faster will use those instead of checking whether a newer go available for download that meets your
constraint. You can change this with the `ignore-local` input below.

A special case value for go-version is `tip` which causes setup-go-faster to install the gotip from source. Be
warned there is nothing fast about this. It takes between 3 and 5 minutes on Ubuntu runners and is even slower
on Windows and MacOS runners.

Go versions aren't really semvers, but they are close enough to use semver constraints for the most part.
There are a some gotchas to watch out for:

- Go doesn't release .0 versions. The first 1.15.x releas is 1.15, not 1.15.0. This means if you have set
  go-version to 1.15, when 1.15.1 is released it won't be used because 1.15 is an exact match. If you want
  any go in the 1.15 family, set go-version to `1.15.x`

- Go's pre-releases are not valid semver. For example the beta for 1.16 is 1.16beta1. This means pre-releases
  need to be explicitely specified.

For those who learn best from examples:

| go-version         | description                                                                                    |
|--------------------|------------------------------------------------------------------------------------------------|
| 1.15.6             | installs 1.15.6                                                                                |
| 1.15.x             | installs the newest go that starts with 1.15                                                   |
| 1.15               | installs go 1.15, nothing newer. You generally do not want this and should use 1.15.x instead. |
| *                  | installs the newest go without any other constraints                                           |
| ^1.15.4            | installs a go that is >= 1.15.4 and < 2                                                        |
| ~1.15.4            | installs a go that is >= 1.15.4 and < 1.16                                                     |
| < 1.15.6 >= 1.15.4 | installs a go that is >= 1.15.4 and < 1.15.6                                                   |
| tip                | installs gotip  from source                                                                    |


### ignore-local

Normally a pre-installed version of go that meets the go-version constraints will be used instead
of checking whether a newer version is available for download. With ignore-local, the
action will always check for a newer version available for download. Set this to any non-empty value
to enable.


## Outputs

### GOCACHE

output of `go env GOCACHE`

### GOMODCACHE

output of `go env GOMODCACHE`

### GOPATH

output of `go env GOPATH`

### GOROOT

output of `go env GOROOT`

### GOTOOLDIR

output of `go env GOTOOLDIR`
