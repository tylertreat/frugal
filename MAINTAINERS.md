## Release

To update the frugal version for release, use the python script provided in the
`scripts` directory. Note: You must be in the root directory of frugal.

Make sure you have the latest master:

```bash
    $ git checkout master && git pull
```

Checkout branch labeled the release version:

```bash
    $ git checkout -b 1_3_1
```

Update version with python update script:

```bash
    $ python scripts/update.py --version=1.3.1
```

Commit and open PR with label:

```
    MSG-116 RELEASE frugal 1.3.1 
```

Where the MSG (or sometimes RM) ticket is the one tracked by MARV. Note: If the
version you are releasing doesn’t match the version tracked by MARV, double
click the version in MARV and change it to the appropriate value before you
open the PR.

Get at least two +1’s, merge and wait for Rosie to tag the release in github.
Download the smithy release artifact from the release build (e.g.  1.3.1)
rename to frugal-1.3.1-linux-amd64 and drag and drop the binary to the release
(e.g. 1.3.1) - note: you’ll have to click `Edit` on the release.
