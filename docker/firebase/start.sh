#!/bin/bash
set -xe

firebase emulators:start --import=./${DATA_DIRECTORY}/export --export-on-exit