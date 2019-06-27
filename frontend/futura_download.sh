#!/bin/bash

DOWNLOAD_PREFIX="http://www.ephifonts.com/downloads"

FONTS=(
    futura-lt.ttf.zip
    futura-lt-bold.ttf.zip
    futura-lt-light.ttf.zip
    futura-lt-condensed.ttf.zip
    futura-lt-oblique.ttf.zip
    futura-lt-book.ttf.zip
    futura-lt-heavy.ttf.zip
    futura-lt-extra-bold.ttf.zip
)

EXTRA_FONTS=(
    futura-lt-condensed-extra-bold-oblique.ttf.zip
    futura-lt-condensed-bold-oblique.ttf.zip
    futura-lt-condensed-oblique.ttf.zip
    futura-lt-condensed-light-oblique.ttf.zip
    futura-lt-extra-bold-oblique.ttf.zip
    futura-lt-book-oblique.ttf.zip
    futura-lt-heavy-oblique.ttf.zip
    futura-lt-condensed-extra-bold.ttf.zip
    futura-lt-light-oblique.ttf.zip
    futura-lt-condensed-bold.ttf.zip
    futura-lt-condensed-light.ttf.zip
    futura-lt-bold-oblique.ttf.zip
)

DESTDIR="resources/public/css/fonts/futura"
TMPDIR=$(mktemp -d)
for font in ${FONTS[*]}; do
    f_loc="${DESTDIR}/${font%.zip}"
    if [ ! -f "${f_loc}" ]; then
	tmploc="${TMPDIR}/${font}"
	curl -s -o "${tmploc}" "${DOWNLOAD_PREFIX}/${font}"
	unzip -q "${tmploc}" -d "${DESTDIR}/"
    else
	echo "${font%.zip} is already downloaded"
    fi
done

rm -rf ${TMPDIR}
