#!/bin/bash
echo "Build"
go run gioui.org/cmd/gogio -appid net.mapopote.mifasol -icon ../docs/logo.png -target android -o assets/mifasolmobile.apk ../cmd/mifasolmobile
adb uninstall net.mapopote.mifasol
adb install assets/mifasolmobile.apk