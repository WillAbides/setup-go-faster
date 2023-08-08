# Setup Go Faster

It's like actions/setup-go but faster.

### A Note About Go 1.21.0

**Use setup-go-faster@v1.9.1 or later** if you want to install Go 1.21.0.

With the release of go1.21.0, the Go team has changed the way they style
dot-zero releases. They used to be styled like `go1.N`, but now they are 
`go1.N.0`. This caused issues with earlier versions of setup-go-faster.

### Faster

Setup-go-faster takes about a third as long as setup-go to install go on a
runner.

These are the median times[^perf-note] for installing go 1.20.7 and tip as of
August 2023.

| runner os    | go-version |    setup-go | setup-go-faster | improvement |
|--------------|------------|------------:|----------------:|------------:|
| ubuntu-20.04 | 1.20.7     |          8s |              3s |          5s |
| macos-11     | 1.20.7     |         34s |             12s |         22s |
| windows-2022 | 1.20.7     |         80s |             12s |         68s |
| ubuntu20.04  | tip        | unsupported |            311s |           ∞ |
| macos-11     | tip        | unsupported |            313s |           ∞ |
| windows-2022 | tip        | unsupported |            576s |           ∞ |

The performance improvements are achieved by:

- The magic of Bash, curl and Perl. Maybe they aren't the most modern, but they
  are a heck of a lot faster than loading nodejs to do some simple version
  checks and downloads.

- Installing to the faster volume on Windows. On windows runners it takes
  significantly longer to write to `C:` vs
  `D:`. Setup-go installs go to `C:`, but setup-go-faster installs to `D:`

- Shortcuts for version checks. Setup-go-faster supports all the same
  pseudo-semver ranges as setup-go, but it is optimized for exact versions (
  like `1.15.7`) and `1.15.x` style ranges. Our version check is faster to begin
  with, but if you use one of those formats you can shave an additional half
  second off the time.

### Install tip

Setup-go-faster will install go tip from source if you set `go-version: tip`.

### Check out the outputs

Look at those outputs. If you want to use GOPATH or GOMODCACHE as input in some
other step, you can just grab it from setup-go-faster\'s output instead of
having to add another step just to set an environment variable.

### What\'s missing?

Just the `stable` input. I don\'t understand what `stable` adds for
actions/setup-go. If you only want stable builds you can set go-version
accordingly. If there is good use case for `stable`, it can be added.

<!--- start generated --->

## Inputs

### go-version

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

- Prior to go1.21, Go doesn't release .0 versions. The first 1.15.x release is 1.15, not 1.15.0. This means if you 
  have set go-version to 1.15, when 1.15.1 is released it won't be used because 1.15 is an exact match. If you
  want any go in the 1.15 family, set go-version to `1.15.x`. For consistency, setup-go-faster@v1 continues to
  handle constraints for post 1.21 the same as pre 1.21. This may change in a future major version.

- Go's pre-releases are not valid semver. For example the beta for 1.16 is 1.16beta1. This means pre-releases
  need to be explicitely specified.

For those who learn best from examples:

| go-version         | description                                                                                    |
|--------------------|------------------------------------------------------------------------------------------------|
| 1.15.6             | installs 1.15.6                                                                                |
| 1.15beta1          | installs 1.15beta1                                                                             |
| 1.15.x             | installs the newest go that starts with 1.15                                                   |
| 1.15               | installs go 1.15, nothing newer. You generally do not want this and should use 1.15.x instead. |
| *                  | installs the newest go without any other constraints                                           |
| ^1.15.4            | installs a go that is >= 1.15.4 and < 2                                                        |
| ~1.15.4            | installs a go that is >= 1.15.4 and < 1.16                                                     |
| < 1.15.6 >= 1.15.4 | installs a go that is >= 1.15.4 and < 1.15.6                                                   |
| tip                | installs gotip  from source                                                                    |


### go-version-file

Path to the go.mod or go.work file.

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
<!--- end generated --->

[^perf-note]: These results come
from [speedrun](https://github.com/WillAbides/test-setup-go-faster/blob/main/.github/workflows/speedrun.yml)
and [speedrun-tip](https://github.com/WillAbides/test-setup-go-faster/blob/main/.github/workflows/speedrun-tip.yml)
from
the [WillAbides/test-setup-go-faster](https://github.com/WillAbides/test-setup-go-faster)
repo.
