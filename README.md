# Setup Go Faster

It's like actions/setup-go but faster.

### Faster

Setup-go-faster takes about a third as long as setup-go to install go on a runner.

These are the median times for installing go 1.15.1 and tip.

| runner os    | go-version |    setup-go | setup-go-faster | improvement |
|--------------|------------|------------:|----------------:|------------:|
| ubuntu-18.04 | 1.15.1     |         11s |              4s |          7s |
| macos-10.15  | 1.15.1     |         20s |              7s |         13s |
| windows-2019 | 1.15.1     |         55s |             18s |         37s |
| ubuntu-18.04 | tip        | unsupported |            136s |           ∞ |
| macos-10.15  | tip        | unsupported |            179s |           ∞ |
| windows-2019 | tip        | unsupported |            260s |           ∞ |

The performance improvements are achieved by:

- Using a simple, composite action instead of loading nodejs.

- Installing to the faster volume on Windows. On windows runners it takes significantly longer to write to `C:` vs
  `D:`. Setup-go installs go to `C:`, but setup-go-faster installs to `D:`

- Shortcuts for version checks. Setup-go-faster supports all the same pseudo-semver ranges as setup-go, but it is
  optimized for exact versions (like `1.15.7`). Our version check is faster to begin with, but
  if you use one of those formats you can shave an additional half second off the time.

### Install tip

Setup-go-faster will install go tip from source if you set `go-version: tip`.

### Check out the outputs

Look at those outputs. If you want to use GOPATH or GOMODCACHE as input in some other step, you can just grab it from
setup-go-faster\'s output instead of having to add another step just to set an environment variable.

### What\'s missing?

Just the `stable` input. I don\'t understand what `stable` adds for actions/setup-go. If you only want stable builds you
can set go-version accordingly. If there is good use case for `stable`, it can be added.

<!--- start generated --->

## Inputs

### go-version

__Required__

The version of go to install. It can be an exact version or a semver constraint like `1.14.x` or `^1.14.4`.
Do not add `go` or `v` to the beginning of the version.

Action runners come with some versions of go pre-installed. If any of those versions meet your semver constraint
setup-go-faster will use those instead of checking whether a newer go available for download that meets your
constraint. You can change this with the `ignore-local` input below.

A special case value for go-version is `tip` which causes setup-go-faster to install the gotip from source. Be
warned there is nothing fast about this. It takes between 3 and 5 minutes on Ubuntu runners and is even slower
on Windows and MacOS runners.

Go versions aren't really semvers, but they are close enough to use semver constraints for the most part.
You do need to remember that go doesn't release .0 versions (the first 1.15.x release is 1.15, not 1.15.0).
This means if you have set go-version to 1.15, when 1.15.1 is released it won't be used because 1.15 is an
exact match. If you want any go in the 1.15 family, set go-version to `1.15.x`.

Prereleases are specified by alphanumeric strings immediately after the minor version (like 1.16beta1). You can
the newest 1.16 including prereleases with `~ 1.16a` because all prereleases will be greater than `a`.

Examples:

| go-version         | description                                                                                    |
|--------------------|------------------------------------------------------------------------------------------------|
| 1.x                | installs the newest go 1 that isn't a prerelease                                               |
| 1.xa               | installs the newest go 1 including prereleases                                                 |
| 1.15.6             | installs 1.15.6                                                                                |
| 1.15beta1          | installs 1.15beta1                                                                             |
| 1.15.x             | installs the newest go that starts with 1.15                                                   |
| 1.15               | installs go 1.15, nothing newer. You generally do not want this and should use 1.15.x instead. |
| ^1.15.4            | installs a go that is >= 1.15.4 and < 2                                                        |
| ~1.15.4            | installs a go that is >= 1.15.4 and < 1.16                                                     |
| >= 1.15.4 < 1.15.6 | installs a go that is >= 1.15.4 and < 1.15.6                                                   |
| tip                | installs gotip  from source                                                                    |


### ignore-local

Normally a pre-installed version of go that meets the go-version constraints will be used instead
of checking whether a newer version is available for download. With ignore-local, the
action will always check for a newer version available for download. Set this to any non-empty value
to enable.


### debug

Set to anything but empty to add debugging information to the logs. It's very noisy and is primarily just
adds 'set -x'.


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
