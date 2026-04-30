BIN=$HOME/bin
TARG=$BIN/Ollie

all:V: $TARG

$TARG:
	mkdir -p $BIN
	go build -o $TARG .

clean:V:
	rm -f $TARG
