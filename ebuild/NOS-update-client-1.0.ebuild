# Copyright 1999-2014 Gentoo Foundation
# Distributed under the terms of the GNU General Public License v2
# $Header: $

EAPI="5"

SLOT="0"

DESCRIPTION="NOS-update-client: To update docker containers"
HOMEPAGE=""
#inherit toolchain-funcs cros-workon systemd
inherit toolchain-funcs systemd

LICENSE=""

if [[ "${PV}" == 9999 ]]; then
    KEYWORDS="~amd64"
else
    #CROS_WORKON_COMMIT="49e0dff2b8529801beb09164729caf96a5b93ef0" # v0.4.6
    KEYWORDS="amd64"
fi

src_unpack() {
	#mkdir NOS-update-client-1.0
	git clone ssh://jfokkan@review.inocybe.com:29418/NOS-update-client
	mv NOS-update-client NOS-update-client-1.0
}

src_compile() {
	./build NOS-update-client
}

src_install() {
	insinto /usr/share/NOS-update-client
	doins ${S}/config.json

	dobin ${S}/bin/${PN}

	systemd_dounit "${FILESDIR}"/${PN}.service
}