package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/wanderer69/frm/internal/bytecode"
	"github.com/wanderer69/frm/internal/compiler"
	"github.com/wanderer69/frm/internal/lexer"
	"github.com/wanderer69/frm/internal/parser"
	"github.com/wanderer69/frm/internal/vm"
)

func TestCompilerI(t *testing.T) {
	data := `функция main() {
    "hello" => ?msg;
    печатать(?msg);
    "привет" => ?msg1;

    фрейм(наименование."фрейм1", сущность."объект") => ?f1;
    фрейм(наименование."фрейм2", сущность."объект") => ?f2;

    если (?msg) {
        печатать("условие сработало");
    };

    для каждого элемента(?f1) => ?x {
        печатать(?x);
    };
    для каждого элемента(?f2) => ?x {
        печатать(?x);
    };
	test1(?msg1);
};

функция test1(?msg1) {
    "hello" => ?msg;
    печатать(?msg);
    печатать(?msg1);
};`

	l := lexer.NewLexer(data)
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	c := compiler.NewCompiler()
	chunk := c.Compile(prog)

	require.NotNil(t, chunk)

	fmt.Printf("%v\r\n", bytecode.PrintChunk(chunk[0]))

	vm := vm.NewVM(chunk)
	require.NoError(t, vm.Run())
}

func TestCompilerII(t *testing.T) {
	data := `функция main() {
    фрейм(наименование."фрейм1", сущность."объект") => ?f1;
    фрейм(наименование."фрейм2", сущность."объект") => ?f2;

	для каждого элемента(?f1) => ?x {
        печатать(?x);
    };
    для каждого элемента(?f2) => ?x {
        печатать(?x);
    };
};`

	l := lexer.NewLexer(data)
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	c := compiler.NewCompiler()
	chunk := c.Compile(prog)

	require.NotNil(t, chunk)

	fmt.Printf("%v\r\n", bytecode.PrintChunk(chunk[0]))

	vm := vm.NewVM(chunk)
	require.NoError(t, vm.Run())
}

func TestCompilerIII(t *testing.T) {
	data := `функция main() {
	3 => ?i1;
	1 => ?i2;
	add(?i1, ?i2) => ?i3;
	печатать(?i3);
	sub(?i1, ?i2) => ?i4;
	печатать(?i4);
	mul(?i4, ?i1) => ?i5;
	печатать(?i5);
	div(?i3, ?i4) => ?i6;
	печатать(?i6);
};`

	l := lexer.NewLexer(data)
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	c := compiler.NewCompiler()
	chunk := c.Compile(prog)

	require.NotNil(t, chunk)

	fmt.Printf("%v\r\n", bytecode.PrintChunk(chunk[0]))

	vm := vm.NewVM(chunk)
	vm.InitIntFunc()
	require.NoError(t, vm.Run())
}

func TestCompilerIV(t *testing.T) {
	//	fA.Slots.Put("parent", &valueType.ValueString{String: "B"})
	//	fB.Slots.Put("child", &valueType.ValueString{String: "C"})

	//	fQ1R1 := frame.NewFrame("query1_rel1")
	//	fQ1R1.Slots.Put("subj", &valueType.ValueString{String: "A"})
	//	fQ1R1.Slots.Put("rel", &valueType.ValueString{String: "parent"})
	//	fQ1R1.Slots.Put("obj", &valueType.ValueString{String: "?X"})

	//	fQ1R2 := frame.NewFrame("query1_rel2")
	//	fQ1R2.Slots.Put("subj", &valueType.ValueString{String: "?X"})
	//	fQ1R2.Slots.Put("rel", &valueType.ValueString{String: "child"})
	//	fQ1R2.Slots.Put("obj", &valueType.ValueString{String: "C"})

	data := `функция main() {
    фрейм(наименование."фрейм1", parent."фрейм2") => ?f1;
    фрейм(наименование."фрейм2", child."фрейм3") => ?f2;
    фрейм(наименование."фрейм3") => ?f3;

    фрейм(наименование."query1_rel1", subj."фрейм1", rel."parent", obj."?X") => ?fq1;
    фрейм(наименование."query1_rel2", subj."?X", rel."child", obj."фрейм3") => ?fq2;
	печатать(?fq1);

	list() => ?l1;
	append(?l1, ?f1) => ?nil;
	append(?l1, ?f2) => ?nil;
	append(?l1, ?f3) => ?nil;

	list() => ?l2;
	append(?l2, ?fq1) => ?nil;
	append(?l2, ?fq2) => ?nil;

	prove(?l1, ?l2) => ?o1;
	печатать(?o1);
};`

	l := lexer.NewLexer(data)
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	c := compiler.NewCompiler()
	chunk := c.Compile(prog)

	require.NotNil(t, chunk)

	fmt.Printf("%v\r\n", bytecode.PrintChunk(chunk[0]))

	vm := vm.NewVM(chunk)
	vm.InitIntFunc()
	require.NoError(t, vm.Run())
}

func TestCompilerV(t *testing.T) {
	data := `функция main() {
    фрейм(наименование."query1_rel1", subj."фрейм1", rel."parent", obj."?X") => ?fq1;
	печатать(?fq1);
};`

	l := lexer.NewLexer(data)
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	c := compiler.NewCompiler()
	chunk := c.Compile(prog)

	require.NotNil(t, chunk)

	fmt.Printf("%v\r\n", bytecode.PrintChunk(chunk[0]))

	vm := vm.NewVM(chunk)
	vm.InitIntFunc()
	require.NoError(t, vm.Run())
}

func TestCompiler1(t *testing.T) {
	l := lexer.NewLexer(test1)
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	c := compiler.NewCompiler()
	chunk := c.Compile(prog)

	require.NotNil(t, chunk)
}

func TestCompiler2(t *testing.T) {
	l := lexer.NewLexer(test2)
	p := parser.NewParser(l)
	prog := p.ParseProgram()
	c := compiler.NewCompiler()
	chunk := c.Compile(prog)

	require.NotNil(t, chunk)
}
