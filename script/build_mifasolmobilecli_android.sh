#!/bin/bash
echo "Build"
cd "$(dirname "$0")"/../cmd/mifasolmobile
fyne package -os android -appID net.mapopote.mifasol -icon ../../docs/logo.png -release -name mifasol
mv mifasol.apk ../../script/assets/
adb uninstall net.mapopote.mifasol
adb install ../../script/assets/mifasol.apk
