Contribution Guidelines
=======================

If you would like to contribute code to this project you can do so
through GitHub by sending a pull request.

If you have a trivial fix or improvement, please create a pull request and
request a review from a [maintainer](MAINTAINERS.md) of this repository.

If you plan to do something more involved, that involves a new feature or
changing functionality, please first create an [issue](#issues) so a discussion of
your idea can happen, avoiding unnecessary work and clarifying implementation.

Development
-----------

This project follows design and development principles stated in:

- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Uber Go Style Guide](https://github.com/uber-go/guide/blob/master/style.md)
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Standard Go Project Layout](https://github.com/golang-standards/project-layout) (partially)
- [Package Oriented Design](https://www.ardanlabs.com/blog/2017/02/package-oriented-design.html)
- [Style Packages](https://rakyll.org/style-packages/)

Documentation
-------------

If you contribute anything that changes the behavior of the library,
document it in the follow places as applicable:

- the code itself, through clear comments and unit tests
- [README](README.md)

This includes new features, additional variants of behavior, and breaking
changes.

Testing
-------

Tests are written using golang's standard testing tools, and are run prior to
the PR being accepted.

Issues
------

For creating an issue:

- **Bugs:** please be as thorough as possible, with steps to recreate the issue
  and any other relevant information.
- **Feature Requests:** please include functionality and use cases.  If this is
  an extension of a current feature, please include whether or not this would
  be a breaking change or how to extend the feature with backwards
  compatibility.
- **Security Vulnerability:** please report it as an [issue](/issues).

If you wish to work on an issue, please assign it to yourself.  If you have any
questions regarding implementation, feel free to ask clarifying questions on
the issue itself.

Pull Requests
-------------

- should be narrowly focused with no more than 3 or 4 logical commits
- when possible, address no more than one issue
- should be reviewable in the GitHub code review tool
- should be linked to any issues it relates to (i.e. issue number after (#) in commit messages or pull request message)
- should conform to idiomatic golang code formatting
