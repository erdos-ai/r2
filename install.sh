#!/bin/sh
OS=`uname -s`
ARCH=`uname -m`

curl -fsSLo $HOME/r2-${OS}-${ARCH}.tar.gz \
	https://github.com/erdos-one/r2/releases/download/v0.2.1-alpha/r2-${OS}-${ARCH}.tar.gz
	
mkdir -p $HOME/r2-v0.2.1-alpha
tar -xzf $HOME/r2-${OS}-${ARCH}.tar.gz -C $HOME/r2-v0.2.1-alpha
chmod +x $HOME/r2-v0.2.1-alpha/r2

if [ "$EUID" -eq 0 ]; then
	mv $HOME/r2-v0.2.1-alpha/r2 /usr/bin/r2
	rm -r $HOME/r2-v0.2.1-alpha
else
	echo "Couldn't add r2 to /usr/bin because script is not running at admin"
	echo "Please run the following commands:"
	echo "  mv $HOME/r2-v0.2.1-alpha/r2-${OS}-${ARCH} /usr/bin/r2"
	echo "  rm -r $HOME/r2-v0.2.1-alpha"
fi
