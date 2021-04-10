%define debug_package %{nil}

Name:    {{{ git_name name="smartbch" }}}
Version: 0.1
Release: {{{ git_version }}}%{?dist}
Summary: A full node client of smartBCH

License:    ASL 2.0
URL:        https://github.com/smartbch/smartbch
VCS:        {{{ git_vcs }}}

Source:     {{{ git_pack }}}

BuildRequires: golang
BuildRequires: git
BuildRequires: gcc-g++

%description
A full node client of smartBCH, an EVM&Web3 compatible sidechain for
Bitcoin Cash.

More information at smartbch.org.

%prep
{{{ git_setup_macro }}}

%build
go build github.com/smartbch/smartbch/cmd/smartbchd

%install
install -D -m 755 ./smartbchd %{buildroot}%{_bindir}/smartbchd

%files
%doc README.md
%license LICENSE
%{_bindir}/smartbchd

%changelog
{{{ git_changelog }}}
