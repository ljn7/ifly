@echo off
setlocal

set "ROOT=%~dp0"
set "PLUGIN=%ROOT%cli\plugin"
set "DIST=%ROOT%dist"
set "MODULE=github.com/ljn7/ifly/cli"
if "%IFLY_REPO_OWNER%"=="" (set "OWNER=ljn7") else (set "OWNER=%IFLY_REPO_OWNER%")
if "%IFLY_REPO_NAME%"=="" (set "REPO=ifly") else (set "REPO=%IFLY_REPO_NAME%")

set /p VERSION=<"%ROOT%VERSION"

if not exist "%PLUGIN%" mkdir "%PLUGIN%"
if not exist "%DIST%" mkdir "%DIST%"

if exist "%PLUGIN%\.claude-plugin" rmdir /s /q "%PLUGIN%\.claude-plugin"
if exist "%PLUGIN%\hooks" rmdir /s /q "%PLUGIN%\hooks"
if exist "%PLUGIN%\skills" rmdir /s /q "%PLUGIN%\skills"
if exist "%PLUGIN%\commands" rmdir /s /q "%PLUGIN%\commands"
if exist "%PLUGIN%\defaults.yaml" del /f /q "%PLUGIN%\defaults.yaml"
if exist "%PLUGIN%\.ifly.yaml.example" del /f /q "%PLUGIN%\.ifly.yaml.example"
if exist "%PLUGIN%\VERSION" del /f /q "%PLUGIN%\VERSION"
if exist "%PLUGIN%\LICENSE" del /f /q "%PLUGIN%\LICENSE"

xcopy /e /i /y "%ROOT%.claude-plugin" "%PLUGIN%\.claude-plugin" >nul || exit /b 1
xcopy /e /i /y "%ROOT%hooks" "%PLUGIN%\hooks" >nul || exit /b 1
xcopy /e /i /y "%ROOT%skills" "%PLUGIN%\skills" >nul || exit /b 1
xcopy /e /i /y "%ROOT%commands" "%PLUGIN%\commands" >nul || exit /b 1
copy /y "%ROOT%defaults.yaml" "%PLUGIN%\defaults.yaml" >nul || exit /b 1
copy /y "%ROOT%.ifly.yaml.example" "%PLUGIN%\.ifly.yaml.example" >nul || exit /b 1
copy /y "%ROOT%VERSION" "%PLUGIN%\VERSION" >nul || exit /b 1
copy /y "%ROOT%LICENSE" "%PLUGIN%\LICENSE" >nul || exit /b 1

pushd "%ROOT%cli" || exit /b 1
set "CGO_ENABLED=0"
go build -ldflags "-s -w -X main.version=%VERSION% -X %MODULE%/cmd.cliVersion=%VERSION% -X %MODULE%/cmd.repoOwner=%OWNER% -X %MODULE%/cmd.repoName=%REPO%" -o "%DIST%\ifly.exe" .
if errorlevel 1 (
  popd
  exit /b 1
)
popd

echo built %DIST%\ifly.exe
