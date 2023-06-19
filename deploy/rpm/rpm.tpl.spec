Summary: indece-monitor agent for linux hosts 
Name: indece-monitor-agent-linux
Version: ${BUILD_VERSION}
Release: 1
URL: https://www.indece.com
Group: System
License: MIT
Packager: indece UG (haftungsbeschränkt)
Requires: glibc
BuildRoot: ${WORK_DIR}
BuildArch: x86_64
AutoReq: no

%description
indece-monitor agent for linux hosts

%install
mkdir -p %{buildroot}/usr/bin/
mkdir -p %{buildroot}/usr/lib/systemd/system/
mkdir -p %{buildroot}/etc/indece-monitor/
cp ${WORK_DIR}/files/indece-monitor-agent-linux %{buildroot}/usr/bin/indece-monitor-agent-linux
cp ${WORK_DIR}/files/indece-monitor-agent-linux.service %{buildroot}/usr/lib/systemd/system/indece-monitor-agent-linux.service
cp ${WORK_DIR}/files/agent-linux.conf %{buildroot}/etc/indece-monitor/agent-linux.conf

%files
/usr/bin/indece-monitor-agent-linux
/usr/lib/systemd/system/indece-monitor-agent-linux.service
%config /etc/indece-monitor/agent-linux.conf

%changelog
* Tue Jun 13 2023 indece <info@indece.com>
- initial version
