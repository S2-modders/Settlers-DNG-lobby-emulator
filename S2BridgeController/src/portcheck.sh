#! /usr/bin/env bash

netstat -tulpn | grep $1 | wc -l
