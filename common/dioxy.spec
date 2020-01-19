################################################################################

# rpmbuilder:relative-pack true

################################################################################

%define _posixroot        /
%define _root             /root
%define _bin              /bin
%define _sbin             /sbin
%define _srv              /srv
%define _home             /home
%define _opt              /opt
%define _lib32            %{_posixroot}lib
%define _lib64            %{_posixroot}lib64
%define _libdir32         %{_prefix}%{_lib32}
%define _libdir64         %{_prefix}%{_lib64}
%define _logdir           %{_localstatedir}/log
%define _rundir           %{_localstatedir}/run
%define _lockdir          %{_localstatedir}/lock/subsys
%define _cachedir         %{_localstatedir}/cache
%define _spooldir         %{_localstatedir}/spool
%define _crondir          %{_sysconfdir}/cron.d
%define _loc_prefix       %{_prefix}/local
%define _loc_exec_prefix  %{_loc_prefix}
%define _loc_bindir       %{_loc_exec_prefix}/bin
%define _loc_libdir       %{_loc_exec_prefix}/%{_lib}
%define _loc_libdir32     %{_loc_exec_prefix}/%{_lib32}
%define _loc_libdir64     %{_loc_exec_prefix}/%{_lib64}
%define _loc_libexecdir   %{_loc_exec_prefix}/libexec
%define _loc_sbindir      %{_loc_exec_prefix}/sbin
%define _loc_bindir       %{_loc_exec_prefix}/bin
%define _loc_datarootdir  %{_loc_prefix}/share
%define _loc_includedir   %{_loc_prefix}/include
%define _loc_mandir       %{_loc_datarootdir}/man
%define _rpmstatedir      %{_sharedstatedir}/rpm-state
%define _pkgconfigdir     %{_libdir}/pkgconfig

################################################################################

%define __ln              %{_bin}/ln
%define __touch           %{_bin}/touch
%define __service         %{_sbin}/service
%define __chkconfig       %{_sbin}/chkconfig
%define __useradd         %{_sbindir}/useradd
%define __groupadd        %{_sbindir}/groupadd
%define __getent          %{_bindir}/getent
%define __systemctl       %{_bindir}/systemctl

################################################################################

%define debug_package     %{nil}

################################################################################

Summary:         Aggregating proxy for MQTT
Name:            dioxy
Version:         1.0.0
Release:         0%{?dist}
Group:           Applications/System
License:         MIT
URL:             https://github.com/gongled/dioxy

Source0:         https://github.com/gongled/%{name}/archive/v%{version}.tar.gz

BuildRoot:       %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

BuildRequires:   golang >= 1.12

Requires:        kaosv >= 2.15

%if 0%{?rhel} >= 7
Requires:        systemd
%endif

Provides:        %{name} = %{version}-%{release}

################################################################################

%description
Aggregating proxy for MQTT broker metrics which represents data in JSON format.

################################################################################

%prep
%setup -q

%build
export GOPATH=$(pwd)

%{__make} %{?_smp_mflags} deps
%{__make} %{?_smp_mflags} %{name}

%install
rm -rf %{buildroot}

install -dm 755 %{buildroot}%{_bindir}
install -dm 755 %{buildroot}%{_sysconfdir}
install -dm 755 %{buildroot}%{_sysconfdir}/logrotate.d
install -dm 755 %{buildroot}%{_initddir}
install -dm 755 %{buildroot}%{_unitdir}
install -dm 755 %{buildroot}%{_logdir}/%{name}
install -dm 755 %{buildroot}%{_rundir}/%{name}

install -pm 755 %{name} \
                %{buildroot}%{_bindir}/

install -pm 644 common/%{name}.knf \
                %{buildroot}%{_sysconfdir}/

install -pm 755 common/%{name}.init \
                %{buildroot}%{_initddir}/%{name}

install -pm 644 common/%{name}.logrotate \
                %{buildroot}%{_sysconfdir}/logrotate.d/%{name}

%if 0%{?rhel} >= 7
install -pDm 644 common/%{name}.service \
                 %{buildroot}%{_unitdir}/
%endif

%clean
rm -rf %{buildroot}

################################################################################

%pre
getent group %{name} >/dev/null || groupadd -r %{name}
getent passwd %{name} >/dev/null || useradd -r -M -g %{name} -s /sbin/nologin %{name}
exit 0

%post
if [[ $1 -eq 1 ]] ; then
    %if 0%{?rhel} <= 6
    %{__chkconfig} --add %{name}
    %else
    %{__systemctl} enable %{name}.service &>/dev/null || :
    %endif
fi

%preun
if [[ $1 -eq 0 ]] ; then
    %if 0%{?rhel} <= 6
    %{__service} %{name} stop &>/dev/null || :
    %{__chkconfig} --del %{service_name}
    %else
    %{__systemctl} --no-reload disable %{name}.service &>/dev/null || :
    %{__systemctl} stop %{name}.service &>/dev/null || :
    %endif
fi

%postun
%if 0%{?rhel} >= 7
if [[ $1 -ge 1 ]] ; then
    %{__systemctl} daemon-reload &>/dev/null || :
fi
%endif

################################################################################

%files
%defattr(-,root,root,-)
%doc LICENSE
%attr(-,%{name},%{name}) %dir %{_logdir}/%{name}
%attr(-,%{name},%{name}) %dir %{_rundir}/%{name}
%config(noreplace) %{_sysconfdir}/%{name}.knf
%config(noreplace) %{_sysconfdir}/logrotate.d/%{name}
%if 0%{?rhel} >= 7
%{_unitdir}/%{name}.service
%endif
%{_initddir}/%{name}
%{_bindir}/%{name}

################################################################################

%changelog
* Sun Jan 19 2020 Gleb Goncharov <inbox@gongled.ru> - 1.0.0-0
- Initial public release
