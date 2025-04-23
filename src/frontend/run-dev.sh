#!/bin/bash

VITE_APP_PROFILE="TEST" npm run serve
echo -ne "\x1b[?25h"
echo -ne "\x1b[0m"