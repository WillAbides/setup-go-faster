# sh-semver

[![NPM Version][npm-image]][npm-url]
[![Build][travis-image]][travis-url]

The semantic versioner for Bourne Shell.

## CLI

```
semver.sh [-r <rule>] [<version>... ]
```

Given a rule and one or many versions, it will return all of the provided versions that satisfy the rule, in sorted order, one per line.

If versions are omitted from the command line, it will read them from STDIN.

If the rule is omitted from the command line, it will simply sort all the provided versions.

## Ranges

Every *version range* contains one or more *sets of comparators*. To satisfy *version range* version must satisfies all *comparators* from at least one set. Sets are separated with two vertical bars ``||``. Rules in each set are separated with whitespaces. Empty set are treated as wildcard comparator ``*``.

### Basic comparators
There are five basic comparators:

* Equal to ``=A.B.C`` (``=`` operator is optional, it may be provided but usually it isn't).
* Less than ``<A.B.C``
* Greater than ``>A.B.C``
* Less than or equal to ``<=A.B.C``
* Greater than or equal to ``>=A.B.C``

### Advanced comparators

##### Wildcards ``*`` ``*.*`` ``*.*.*``
Wildcard comparators can be satisfied by any version, even prerelease one.

Instead of ``*`` character can be used ``x`` or ``X``.

##### Wildcard ranges ``A.*.*`` ``A.B.*``
Wildcard ranges can be satisfied when only part of version is matching. In contrast to wildcards, wildcard ranges cannot be satisfied by prerelease versions.

* ``A.*.*`` is satisfied when majors are the same,
* ``A.B.*`` is satisfied when majors and minors are the same.

The special character (``*``, ``x`` or ``X``) is optional.

###### Examples
* ``1.2`` := ``>=1.2.0 <1.3.0``
* ``5.*.*`` := ``>=5.0.0 <6.0.0``

##### Caret ranges ``^A.B.C``
Matches to compatible versions.

###### Examples
* ``^1.2.3`` := ``>=1.2.3 <2.0.0``
* ``^0.1.2`` := ``>=0.1.2 <0.2.0``
* ``^0.0.1`` := ``>=0.0.1 <0.0.2``

##### Tilde ranges ``~A.B.C`` ``~A.B`` ``~A``
If patch version is specified tilde ranges matches to all greater or equal versions with the same minor version. Otherwise is equivalent of ``A.B.*`` when minor version is specified or ``A.*.*`` when not.

##### Hyphen ranges ``A.B.C - X.Y.Z``
Hyphen range ``A.B.C - X.Y.Z`` equivalent of ``>=A.B.C <=X.Y.Z``.

### Prerelease versions
Prerelease versions can satisfy comparators set only when have the same minor major and patch numbers as at least one of comparators.


[npm-image]: https://img.shields.io/npm/v/sh-semver.svg
[npm-url]: https://npmjs.org/package/sh-semver.sh
[travis-image]: https://travis-ci.org/qzb/sh-semver.svg?branch=master
[travis-url]: https://travis-ci.org/qzb/sh-semver
