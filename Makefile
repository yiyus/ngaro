include $(GOROOT)/src/Make.inc

TARG=ngaro
GOFILES=ngaro.go dev.go file.go img.go

CLEANFILES+=gonga

include $(GOROOT)/src/Make.pkg

main.$O: main.go package
	$(GC) -I_obj $<

gonga: main.$O
	$(LD) -L_obj -o $@ $<
