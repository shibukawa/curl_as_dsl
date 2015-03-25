package form2curl

import (
	. "gopkg.in/check.v1"
)

type Form2CurlTest struct{}

var _ = Suite(&Form2CurlTest{})

func (s *Form2CurlTest) Test_FormIsGone(c *C) {
	src := `<pre>There is no form</pre>`
	form, err := CreateFormFromString(src)
	c.Assert(err, NotNil)
	c.Assert(form, IsNil)
}

func (s *Form2CurlTest) Test_EmptyForm(c *C) {
	src := `
	<form></form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(form.Method, Equals, "GET")
	c.Check(form.Action, Equals, "")
	c.Check(len(form.Inputs), Equals, 0)
}

func (s *Form2CurlTest) Test_FormParameters(c *C) {
	src := `
	<form method="POST" action="http://localhost:18888" enctype="multipart/form-data" accept-charset="UTF-8" accept="text/html;text/plain"></form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(form.Method, Equals, "POST")
	c.Check(form.Action, Equals, "http://localhost:18888")
	c.Check(form.EncType, Equals, "multipart/form-data")
	c.Check(form.AcceptCharset, Equals, "UTF-8")
	c.Check(form.Accept, Equals, "text/html;text/plain")
}

func (s *Form2CurlTest) Test_FormParameterErrorTest_1(c *C) {
	// PUT is invalid
	src := `
	<form method="PUT" action="http://localhost:18888"></form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Warnings), Equals, 1)
}

func (s *Form2CurlTest) Test_FormParameterErrorTest_2(c *C) {
	// it should be http/https or path style url
	src := `
	<form action="ftp://localhost:18888"></form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Warnings), Equals, 1)
}

func (s *Form2CurlTest) Test_FormParameterErrorTest_3(c *C) {
	// enctype is only for POST
	src := `
	<form method="GET" enctype="application/x-www-form-urlencoded"></form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Warnings), Equals, 1)
}

func (s *Form2CurlTest) Test_FormParameterErrorTest_4(c *C) {
	src := `
	<form method="POST" enctype="application/x-www-form-urlencoded"></form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(form.EncType, Equals, "application/x-www-form-urlencoded")
}

func (s *Form2CurlTest) Test_FormParameterErrorTest_5(c *C) {
	src := `
	<form method="POST" enctype="multipart/form-data"></form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(form.EncType, Equals, "multipart/form-data")
}

func (s *Form2CurlTest) Test_FormParameterErrorTest_6(c *C) {
	src := `
	<form method="POST" enctype="text/plain"></form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(form.EncType, Equals, "text/plain")
}

func (s *Form2CurlTest) Test_FormParameterErrorTest_7(c *C) {
	// enctype accepts only above three types
	src := `
	<form method="POST" enctype="application/json"></form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Warnings), Equals, 1)
}

func (s *Form2CurlTest) Test_TextInput(c *C) {
	src := `
	<form>
		<input name="single_input" value="initial value">
		<input type="text" name="single_input2" value="">
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Inputs), Equals, 2)
	c.Check(form.Inputs[0].Type, Equals, "text")
	c.Check(form.Inputs[0].Name, Equals, "single_input")
	c.Check(form.Inputs[0].Value, Equals, "initial value")
	c.Check(form.Inputs[0].ValueString(true), Equals, "initial value")
	c.Check(form.Inputs[1].Type, Equals, "text")
	c.Check(form.Inputs[1].Name, Equals, "single_input2")
	c.Check(form.Inputs[1].Value, Equals, "")
	c.Check(form.Inputs[1].ValueString(true), Equals, "dummy text")
	c.Check(form.Inputs[1].CanOmit(), Equals, false)
}

func (s *Form2CurlTest) Test_TextInputWithDataList_1(c *C) {
	src := `
	<form>
		<input type="text" name="browser" list="browsers">
		<datalist id="browsers">
			<option value="Internet Explorer">
			<option value="Firefox">
			<option value="Chrome">
			<option value="Opera">
			<option value="Safari">
		</datalist>
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Inputs), Equals, 1)
	c.Check(form.Inputs[0].Type, Equals, "text")
	c.Check(form.Inputs[0].Name, Equals, "browser")
	c.Check(form.Inputs[0].DataListName, Equals, "browsers")
	c.Check(form.Inputs[0].ValueString(true), Equals, "Internet Explorer")
	c.Check(len(form.Inputs[0].Selections), Equals, 5)
}

func (s *Form2CurlTest) Test_TextInputWithDataList_2(c *C) {
	// specify not exist list name
	src := `
	<form>
		<input type="text" name="browser" list="operating_system">
		<datalist id="browsers">
			<option value="Internet Explorer">
			<option value="Firefox">
			<option value="Chrome">
			<option value="Opera">
			<option value="Safari">
		</datalist>
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Inputs), Equals, 1)
	c.Check(form.Inputs[0].Type, Equals, "text")
	c.Check(form.Inputs[0].Name, Equals, "browser")
	c.Check(form.Inputs[0].DataListName, Equals, "operating_system")
	c.Check(len(form.Warnings), Equals, 1)
}

func (s *Form2CurlTest) Test_RadioButton_1(c *C) {
	src := `
	<form>
		<input type="radio" name="radio1" value="opt1">opt1</input>
		<input type="radio" name="radio1" value="opt2" checked="checked">opt2</input>
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Inputs), Equals, 1)
	c.Check(form.Inputs[0].Type, Equals, "radio")
	c.Check(form.Inputs[0].Name, Equals, "radio1")
	c.Check(form.Inputs[0].Value, Equals, "opt2")
	c.Check(form.Inputs[0].ValueString(true), Equals, "opt2")
	c.Check(form.Inputs[0].Selections[0], Equals, "opt1")
	c.Check(form.Inputs[0].Selections[1], Equals, "opt2")
}

func (s *Form2CurlTest) Test_RadioButton_2(c *C) {
	src := `
	<form>
		<input type="radio" name="radio1" value="opt1">opt1</input>
		<input type="radio" name="radio1" value="opt2">opt2</input>
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Inputs), Equals, 1)
	c.Check(form.Inputs[0].Type, Equals, "radio")
	c.Check(form.Inputs[0].Name, Equals, "radio1")
	c.Check(form.Inputs[0].Value, Equals, "")
	c.Check(form.Inputs[0].ValueString(true), Equals, "opt1")
	c.Check(form.Inputs[0].Selections[0], Equals, "opt1")
	c.Check(form.Inputs[0].Selections[1], Equals, "opt2")
}

func (s *Form2CurlTest) Test_Select_1(c *C) {
	src := `
	<form>
		<select name="blood">
			<option value="A">A</option>
			<option value="B">B</option>
			<option value="O">O</option>
			<option value="AB" selected="selected">AB</option>
		</select>
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Inputs), Equals, 1)
	c.Check(form.Inputs[0].Type, Equals, "select")
	c.Check(form.Inputs[0].Name, Equals, "blood")
	c.Check(form.Inputs[0].Value, Equals, "AB")
	c.Check(form.Inputs[0].ValueString(true), Equals, "AB")
	c.Check(form.Inputs[0].Selections[0], Equals, "A")
	c.Check(form.Inputs[0].Selections[1], Equals, "B")
	c.Check(form.Inputs[0].Selections[2], Equals, "O")
	c.Check(form.Inputs[0].Selections[3], Equals, "AB")
}

func (s *Form2CurlTest) Test_Select_2(c *C) {
	src := `
	<form>
		<select name="blood">
			<option value="A">A</option>
			<option value="B">B</option>
			<option value="O">O</option>
			<option value="AB">AB</option>
		</select>
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Inputs), Equals, 1)
	c.Check(form.Inputs[0].Type, Equals, "select")
	c.Check(form.Inputs[0].Name, Equals, "blood")
	c.Check(form.Inputs[0].Value, Equals, "")
	c.Check(form.Inputs[0].ValueString(true), Equals, "A")
	c.Check(form.Inputs[0].Selections[0], Equals, "A")
	c.Check(form.Inputs[0].Selections[1], Equals, "B")
	c.Check(form.Inputs[0].Selections[2], Equals, "O")
	c.Check(form.Inputs[0].Selections[3], Equals, "AB")
}

func (s *Form2CurlTest) Test_TextArea(c *C) {
	src := `
	<form>
		<textarea name="textarea1">initial value</textarea>
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Inputs), Equals, 1)
	c.Check(form.Inputs[0].Type, Equals, "textarea")
	c.Check(form.Inputs[0].Name, Equals, "textarea1")
	c.Check(form.Inputs[0].Value, Equals, "initial value")
}

func (s *Form2CurlTest) Test_CheckBox(c *C) {
	src := `
	<form>
		<input name="checkbox1" type="checkbox" value="ng">
		<input name="checkbox2" type="checkbox" value="ok" checked="checked">
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Inputs), Equals, 2)
	c.Check(form.Inputs[0].Type, Equals, "checkbox")
	c.Check(form.Inputs[0].Name, Equals, "checkbox1")
	c.Check(form.Inputs[0].Value, Equals, "")
	c.Check(form.Inputs[0].Selected, Equals, false)
	c.Check(form.Inputs[0].ValueString(true), Equals, "ng")
	c.Check(form.Inputs[1].Type, Equals, "checkbox")
	c.Check(form.Inputs[1].Name, Equals, "checkbox2")
	c.Check(form.Inputs[1].Value, Equals, "ok")
	c.Check(form.Inputs[1].Selected, Equals, true)
	c.Check(form.Inputs[1].ValueString(true), Equals, "ok")
}

func (s *Form2CurlTest) Test_Range(c *C) {
	/*
		Default Values:
			min: 0
			max: 100
			value: min + (max-min)/2, or min if max is less than min
			step: 1
	*/
	src := `
	<form>
		<input name="range1" type="range">50
		<input name="range2" type="range" min="10" max="20">15
		<input name="range3" type="range" min="20" max="10">20
		<input name="range4" type="range" min="10" max="13">12
		<input name="range5" type="range" min="10" max="20" value="15">15
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Inputs), Equals, 5)
	c.Check(form.Inputs[0].Type, Equals, "range")
	c.Check(form.Inputs[0].Name, Equals, "range1")
	c.Check(form.Inputs[0].Value, Equals, "50")
	c.Check(form.Inputs[1].Type, Equals, "range")
	c.Check(form.Inputs[1].Name, Equals, "range2")
	c.Check(form.Inputs[1].Value, Equals, "15")
	c.Check(form.Inputs[2].Type, Equals, "range")
	c.Check(form.Inputs[2].Name, Equals, "range3")
	c.Check(form.Inputs[2].Value, Equals, "20")
	c.Check(form.Inputs[3].Type, Equals, "range")
	c.Check(form.Inputs[3].Name, Equals, "range4")
	c.Check(form.Inputs[3].Value, Equals, "12")
	c.Check(form.Inputs[4].Type, Equals, "range")
	c.Check(form.Inputs[4].Name, Equals, "range5")
	c.Check(form.Inputs[4].Value, Equals, "15")
}

func (s *Form2CurlTest) Test_FileInput_1(c *C) {
	// error when method is get
	src := `
	<form>
		<input type="file" name="sentfile">
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Warnings), Equals, 2)
}

func (s *Form2CurlTest) Test_FileInput_2(c *C) {
	// error when enctype is not multipart/form-data
	src := `
	<form method="post" enctype="application/x-www-form-urlencoded">
		<input type="file" name="sentfile">
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Warnings), Equals, 1)
}

func (s *Form2CurlTest) Test_FileInput_3(c *C) {
	src := `
	<form method="post" enctype="multipart/form-data">
		<input type="file" name="sentfile">
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Warnings), Equals, 0)
	c.Check(len(form.Inputs), Equals, 1)
	c.Check(form.Inputs[0].Type, Equals, "file")
	c.Check(form.Inputs[0].Name, Equals, "sentfile")
	c.Check(form.Inputs[0].ValueString(true), Equals, "@test.txt")
}

func (s *Form2CurlTest) Test_Number(c *C) {
	src := `
	<form method="post">
		<input type="number" name="regular">
		<input type="number" name="preset" value="20">
		<input type="number" name="range1" min="10">
		<input type="number" name="range2" max="20">
		<input type="number" name="range3" min="10" max="20">
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Warnings), Equals, 0)
	c.Check(len(form.Inputs), Equals, 5)
	c.Check(form.Inputs[0].Type, Equals, "number")
	c.Check(form.Inputs[0].Name, Equals, "regular")
	c.Check(form.Inputs[0].Value, Equals, "")
	c.Check(form.Inputs[0].ValueString(true), Equals, "1234567")
	c.Check(form.Inputs[1].Type, Equals, "number")
	c.Check(form.Inputs[1].Name, Equals, "preset")
	c.Check(form.Inputs[1].Value, Equals, "20")
	c.Check(form.Inputs[1].ValueString(true), Equals, "20")
	c.Check(form.Inputs[2].Type, Equals, "number")
	c.Check(form.Inputs[2].Name, Equals, "range1")
	c.Check(form.Inputs[2].Value, Equals, "")
	c.Check(form.Inputs[2].ValueString(true), Equals, "10")
	c.Check(form.Inputs[3].Type, Equals, "number")
	c.Check(form.Inputs[3].Name, Equals, "range2")
	c.Check(form.Inputs[3].Value, Equals, "")
	c.Check(form.Inputs[3].ValueString(true), Equals, "20")
	c.Check(form.Inputs[4].Type, Equals, "number")
	c.Check(form.Inputs[4].Name, Equals, "range3")
	c.Check(form.Inputs[4].Value, Equals, "")
	c.Check(form.Inputs[4].ValueString(true), Equals, "15")
}

func (s *Form2CurlTest) Test_OtherAcceptableTypes(c *C) {
	src := `
	<form>
		<input name="1" type="color">
		<input name="2" type="date">
		<input name="3" type="datetime">
		<input name="4" type="datetime-local">
		<input name="5" type="email">
		<input name="6" type="hidden">
		<input name="7" type="month">
		<input name="8" type="password">
		<input name="9" type="search">
		<input name="10" type="tel">
		<input name="11" type="time">
		<input name="12" type="url">
		<input name="13" type="week">
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Warnings), Equals, 0)
	c.Check(len(form.Inputs), Equals, 13)
	c.Check(form.Inputs[0].ValueString(true), Equals, "#0592D0")
	c.Check(form.Inputs[1].ValueString(true), Equals, "2015-03-23")
	c.Check(form.Inputs[6].ValueString(true), Equals, "2015-03")
	c.Check(form.Inputs[9].ValueString(true), Equals, "+1-202-555-0145")
	c.Check(form.Inputs[10].ValueString(true), Equals, "12:30")
	c.Check(form.Inputs[12].ValueString(true), Equals, "2015-W04")
}

func (s *Form2CurlTest) Test_IgnoreTypes(c *C) {
	src := `
	<form>
		<input name="1" type="button">
		<input name="2" type="image">
		<input name="3" type="reset">
		<input name="4" type="submit">
	</form>
	`
	form, err := CreateFormFromString(src)
	c.Check(err, Equals, nil)
	c.Check(len(form.Inputs), Equals, 0)
	c.Check(len(form.Warnings), Equals, 0)
}
