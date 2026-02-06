# go-binlog-sync

## Overview

go-binlog-sync is a binlog sync tool. It reads binlog from source mysql and syncs to destination mysql.

- cmd/gbs is the main entry point.
- pkg/logger is the logger package.
- pkg/context contains the configuration context, all configurations are loaded from this context.


## Technology Stack

- Go: 1.25

## Core Rules