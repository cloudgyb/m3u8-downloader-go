Unicode true

####
## Please note: Template replacements don't work in this file. They are provided with default defines like
## mentioned underneath.
## If the keyword is not defined, "wails_tools.nsh" will populate them with the values from ProjectInfo.
## If they are defined here, "wails_tools.nsh" will not touch them. This allows to use this project.nsi manually
## from outside of Wails for debugging and development of the installer.
##
## For development first make a wails nsis build to populate the "wails_tools.nsh":
## > wails build --target windows/amd64 --nsis
## Then you can call makensis on this file with specifying the path to your binary:
## For a AMD64 only installer:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app.exe
## For a ARM64 only installer:
## > makensis -DARG_WAILS_ARM64_BINARY=..\..\bin\app.exe
## For a installer with both architectures:
## > makensis -DARG_WAILS_AMD64_BINARY=..\..\bin\app-amd64.exe -DARG_WAILS_ARM64_BINARY=..\..\bin\app-arm64.exe
####
## The following information is taken from the ProjectInfo file, but they can be overwritten here.
####
## !define INFO_PROJECTNAME    "MyProject" # Default "{{.Name}}"
## !define INFO_COMPANYNAME    "MyCompany" # Default "{{.Info.CompanyName}}"
## !define INFO_PRODUCTNAME    "MyProduct" # Default "{{.Info.ProductName}}"
## !define INFO_PRODUCTVERSION "1.0.0"     # Default "{{.Info.ProductVersion}}"
## !define INFO_COPYRIGHT      "Copyright" # Default "{{.Info.Copyright}}"
###
## !define PRODUCT_EXECUTABLE  "Application.exe"      # Default "${INFO_PROJECTNAME}.exe"
## !define UNINST_KEY_NAME     "UninstKeyInRegistry"  # Default "${INFO_COMPANYNAME}${INFO_PRODUCTNAME}"
####
## !define REQUEST_EXECUTION_LEVEL "admin"            # Default "admin"  see also https://nsis.sourceforge.io/Docs/Chapter4.html
####
## Include the wails tools
####
!include "wails_tools.nsh"

# The version information for this two must consist of 4 parts
VIProductVersion "${INFO_PRODUCTVERSION}.0"
VIFileVersion    "${INFO_PRODUCTVERSION}.0"

VIAddVersionKey "CompanyName"     "${INFO_COMPANYNAME}"
VIAddVersionKey "FileDescription" "${INFO_PRODUCTNAME} Installer"
VIAddVersionKey "ProductVersion"  "${INFO_PRODUCTVERSION}"
VIAddVersionKey "FileVersion"     "${INFO_PRODUCTVERSION}"
VIAddVersionKey "LegalCopyright"  "${INFO_COPYRIGHT}"
VIAddVersionKey "ProductName"     "${INFO_PRODUCTNAME}"

# Enable HiDPI support. https://nsis.sourceforge.io/Reference/ManifestDPIAware
ManifestDPIAware true

!include "MUI2.nsh"
!include "StrFunc.nsh"
${StrLoc} # Defines the StrLoc function used to detect a running process.

!define MUI_ICON "..\icon.ico"
!define MUI_UNICON "..\icon.ico"

# Left image on the welcome / finish pages (164x314).
!define MUI_WELCOMEFINISHPAGE_BITMAP "resources\welcome.bmp"

!define MUI_INSTFILESPAGE_PROGRESSBAR_COLORED "2563eb" # Brand-colored progress bar.
!define MUI_FINISHPAGE_NOAUTOCLOSE # Wait on the INSTFILES page so the user can take a look into the details of the installation steps
!define MUI_ABORTWARNING # This will warn the user if they exit from the installer.
!define MUI_FINISHPAGE_RUN "$INSTDIR\${PRODUCT_EXECUTABLE}" # Show a checkbox to launch the app after install.
!define MUI_FINISHPAGE_RUN_TEXT "运行 ${INFO_PRODUCTNAME}"

!insertmacro MUI_PAGE_WELCOME # Welcome to the installer page.
!insertmacro MUI_PAGE_LICENSE "resources\eula.txt" # EULA page; the user must accept before install.
!insertmacro MUI_PAGE_DIRECTORY # In which folder install page.
!insertmacro MUI_PAGE_INSTFILES # Installing page.
!insertmacro MUI_PAGE_FINISH # Finished installation page.

!insertmacro MUI_UNPAGE_INSTFILES # Uinstalling page

!insertmacro MUI_LANGUAGE "SimpChinese" # Set the Language of the installer

# Fix: the license-page hint referenced "我同意(I)" while the accept button
# reads "我接受(I)"; align the hint with the button label.
LangString MUI_INNERTEXT_LICENSE_BOTTOM ${LANG_SIMPCHINESE} "如果你接受许可证的条款，请点击 [我接受(I)] 继续安装。你必须在接受后才能安装 $(^NameDA) 。"

## The following two statements can be used to sign the installer and the uninstaller. The path to the binaries are provided in %1
#!uninstfinalize 'signtool --file "%1"'
#!finalize 'signtool --file "%1"'

Name "${INFO_PRODUCTNAME}"
OutFile "..\..\bin\${INFO_PROJECTNAME}-${ARCH}-installer.exe" # Name of the installer's file.
!ifdef WAILS_INSTALL_SCOPE
  !if "${WAILS_INSTALL_SCOPE}" == "user"
    InstallDir "$LOCALAPPDATA\Programs\${INFO_PRODUCTNAME}"
  !else
    InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
  !endif
!else
  InstallDir "$PROGRAMFILES64\${INFO_COMPANYNAME}\${INFO_PRODUCTNAME}"
!endif # Default installing folder ($PROGRAMFILES is Program Files folder).
ShowInstDetails show # This will always show the installation details.
BrandingText "M3U8 下载器 · CloudGYB" # Footer text (bottom-left).

Function .onInit
   !insertmacro wails.checkArchitecture

   # Default to the previous install directory so an existing install is
   # overwritten in place instead of reinstalled side by side.
   Call GetPreviousInstallDir

   # If the app is already running, ask the user to terminate it first.
   Call CheckAndTerminateRunningApp
FunctionEnd

# Reads the previous version's install directory from the registry and uses
# it as the default $INSTDIR.
Function GetPreviousInstallDir
    SetRegView 64
    ReadRegStr $R0 HKLM "${UNINST_KEY}" "UninstallString"
    ${If} $R0 == ""
        ReadRegStr $R0 HKCU "${UNINST_KEY}" "UninstallString"
    ${EndIf}
    ${If} $R0 != ""
        # $R0 is quoted, e.g. "C:\Program Files\...\uninstall.exe"
        StrCpy $R1 $R0 "" 1    # drop the leading quote
        StrCpy $R1 $R1 -1      # drop the trailing quote
        ${GetParent} "$R1" $R2 # $R2 = the previous install directory
        StrCpy $INSTDIR $R2
    ${EndIf}
FunctionEnd

# If ${PRODUCT_EXECUTABLE} is already running, offer to terminate it before
# overwriting, otherwise abort the install.
Function CheckAndTerminateRunningApp
    nsExec::ExecToStack 'tasklist /FI "IMAGENAME eq ${PRODUCT_EXECUTABLE}" /NH'
    Pop $0  # return code
    Pop $1  # stdout
    ${StrLoc} $2 $1 "${PRODUCT_EXECUTABLE}" ">"
    ${If} $2 != ""
        MessageBox MB_YESNO|MB_ICONEXCLAMATION "${INFO_PRODUCTNAME} 正在运行，是否终止它并继续安装？" IDYES terminate IDNO cancel
        Goto done
        terminate:
            nsExec::Exec 'taskkill /F /IM ${PRODUCT_EXECUTABLE}'
            Sleep 1000
            Goto done
        cancel:
            Abort
        done:
    ${EndIf}
FunctionEnd

Section
    !insertmacro wails.setShellContext

    !insertmacro wails.webview2runtime

    SetOutPath $INSTDIR

    !insertmacro wails.files

    # Bundle ffmpeg.exe for MP4 remuxing when present next to the built binary.
    File /nonfatal "..\..\bin\ffmpeg.exe"

    CreateShortcut "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"
    CreateShortCut "$DESKTOP\${INFO_PRODUCTNAME}.lnk" "$INSTDIR\${PRODUCT_EXECUTABLE}"

    !insertmacro wails.associateFiles
    !insertmacro wails.associateCustomProtocols

    !insertmacro wails.writeUninstaller
SectionEnd

Section "uninstall"
    !insertmacro wails.setShellContext

    RMDir /r "$AppData\${PRODUCT_EXECUTABLE}" # Remove the WebView2 DataPath

    RMDir /r $INSTDIR

    Delete "$SMPROGRAMS\${INFO_PRODUCTNAME}.lnk"
    Delete "$DESKTOP\${INFO_PRODUCTNAME}.lnk"

    !insertmacro wails.unassociateFiles
    !insertmacro wails.unassociateCustomProtocols

    !insertmacro wails.deleteUninstaller
SectionEnd
