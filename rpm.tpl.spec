Summary: indece-monitor agent for linux hosts 
Name: indece-monitor-checker-linux
Version: ${BUILD_VERSION}
Release: 1
URL: https://www.indece.com
Group: System
License: MIT
Packager: indece UG (haftungsbeschränkt)
Requires: bash
BuildRoot: ${WORK_DIR}

%description
indece-monitor agent for linux hosts

%install
mkdir -p %{buildroot}/usr/bin/
cp ${WORK_DIR}/files/indece-monitor-agent-linux %{buildroot}/usr/bin/indece-monitor-agent-linux

%files
/usr/bin/indece-monitor-agent-linux

%changelog
* Tue Jun 13 2023 indece <info@indece.com>
- initial version
