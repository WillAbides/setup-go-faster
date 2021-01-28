### Faster

Setup-go-faster takes about a third as long as setup-go to install go on a runner.

These are the median times for installing go 1.15.1.

| runner os    | setup-go | setup-go-faster | improvement |
|--------------|---------:|----------------:|------------:|
| ubuntu-18.04 |      11s |              4s |          7s |
| macos-10.15  |      20s |              7s |         13s |
| windows-2019 |      55s |             18s |         37s |

When using a pre-installed version of go, setup-go-faster will be done less than a second vs 1-2 seconds for setup-go.

The performance improvements are achieved by:

- The magic of Bash, curl and Perl. Maybe they aren't the most modern, but they are a heck of a lock faster than loading
  nodejs to do some simple version checks and downloads.

- Installing to the faster volume on Windows. On windows runners it takes significantly longer to write to `C:` vs
  `D:`. Setup-go installs go to `C:`, but setup-go-faster installs to `D:`

- Shortcuts for version checks. Setup-go-faster supports all the same pseudo-semver ranges as setup-go, but it is
  optimized for exact versions (like `1.15.7`) and `1.15.x` style ranges. Our version check is faster to begin with, but
  if you use one of those formats you can shave an additional half second off the time.

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

Just the `stable` input. I don\'t understand what `stable` adds for actions/setup-go. If you only want stable builds you
can set go-version accordingly. If there is good use case for `stable`, it can be added.

