@ECHO OFF

@SET GOOS=linux
@SET GOARCH=amd64
@SET FILENAME=scaleconnect_linux_amd64
go build -ldflags "-s -w" -trimpath -o %FILENAME% && upx --best --lzma %FILENAME%

@SET GOOS=linux
@SET GOARCH=arm64
@SET FILENAME=scaleconnect_linux_arm64
go build -ldflags "-s -w" -trimpath -o %FILENAME% && upx --best --lzma %FILENAME%

@SET GOOS=linux
@SET GOARCH=arm
@SET GOARM=7
@SET FILENAME=scaleconnect_linux_arm
go build -ldflags "-s -w" -trimpath -o %FILENAME% && upx --best --lzma %FILENAME%

@SET GOOS=windows
@SET GOARCH=amd64
@SET FILENAME=scaleconnect_win64.zip
go build -ldflags "-s -w" -trimpath -o scaleconnect.exe && 7z a -mx9 -sdel %FILENAME% scaleconnect.exe && 7z a %FILENAME% %~dp0scaleconnect.yaml

@SET GOOS=darwin
@SET GOARCH=amd64
@SET FILENAME=scaleconnect_mac_amd64.zip
go build -ldflags "-s -w" -trimpath -o scaleconnect && python %~dp0zip.py %FILENAME% scaleconnect && 7z a %FILENAME% %~dp0scaleconnect.yaml

@SET GOOS=darwin
@SET GOARCH=arm64
@SET FILENAME=scaleconnect_mac_arm64.zip
go build -ldflags "-s -w" -trimpath -o scaleconnect && python %~dp0zip.py %FILENAME% scaleconnect && 7z a %FILENAME% %~dp0scaleconnect.yaml
