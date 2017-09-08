# log: Simple wrapper for logrus

`logrus` by default outputs everything to `stderr` which makes Infof, Debugf messages appear to be error messages.

This package provides only the following:

* Sends Infof, Debugf messages to `stdout`.
* Sends Errorf messages to go `stderr`.
* Allow the set the log level.

All other things are not supported because I only care about Infof, Debugf and Errorf.

Example: https://github.com/leodotcloud/log-example
