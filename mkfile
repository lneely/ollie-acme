BIN=$HOME/bin
TARG=$BIN/Ollie

all:V:
	mkdir -p $BIN
	go build -o $TARG .

clean:V:
	rm -f $TARG
