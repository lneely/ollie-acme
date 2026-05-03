BIN=$HOME/bin
TARG=$BIN/Ollie

all:V:
	mkdir -p $BIN
	go build -o $TARG .
	cp scripts/Kmpl $BIN/Kmpl
	chmod +x $BIN/Kmpl

clean:V:
	rm -f $TARG $BIN/Kmpl
