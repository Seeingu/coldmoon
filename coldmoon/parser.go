package coldmoon

import (
	"strings"

	"github.com/samber/lo"
)

type Parser struct {
	SourceText              string
	tokenizer               *Tokenizer
	inFunctionBody          bool
	inFormalParameters      bool
	inClassBody             bool
	inMethodDefinition      bool
	inClassConstructor      bool
	inIteration             bool
	inBreakable             bool
	inModule                bool
	callExpressionForbidden bool
	ctx                     ParserContext
}

func NewParser(sourceText string, ctx ParserContext) *Parser {
	return &Parser{
		SourceText: sourceText,
		tokenizer:  NewTokenizer(sourceText),
		ctx:        ctx,
	}
}

func (p *Parser) Parse() *Script {
	return &Script{
		StatementList: p.ParseNode(),
	}
}

func (p *Parser) ParseModule() *Module {
	inModule := p.inModule
	p.inModule = true
	defer func() {
		p.inModule = inModule
	}()

	return p.module()
}

func (p *Parser) module() *Module {
	moduleItemList := p.moduleItemList()
	return &Module{
		ModuleItemList: moduleItemList,
	}
}

func (p *Parser) moduleItemList() ModuleItemList {
	var items ModuleItemList
	for {
		item := p.moduleItem()
		if item == nil {
			break
		}
		items = append(items, item)
	}
	return items
}

func (p *Parser) moduleItem() ModuleItem {
	t := p.tokenizer.CurrentToken
	switch t.Type {
	case TImport:
		return p.importDeclaration()
	case TExport:
		return p.exportDeclaration()
	case TEOF:
		return nil
	default:
		item := p.statementListItem()
		return &ModuleItemStatementListItem{
			StatementListItem: item,
		}
	}
}

func (p *Parser) exportDeclaration() *ModuleItemExportDeclaration {
	m := &ModuleItemExportDeclaration{}
	p.tokenizer.MustMatch(TExport)
	if p.tokenizer.Match(TDefault) {
		if d, ok := parserRecoverOk(p, p.hoistableDeclaration); ok {
			m.DefaultHoistableDeclaration = d
		} else if c, ok := parserRecoverOk(p, p.classDeclaration); ok {
			m.DefaultClassDeclaration = c
		} else if e, ok := parserRecoverOk[Expression](p, func() Expression {
			return p.expression(p.acceptContextLowest())
		}); ok {
			m.DefaultExpression = e
		} else {
			panic("exportDeclaration: expected hoistable declaration, class declaration or expression")
		}
	} else if e, ok := parserRecoverOk(p, p.exportFrom); ok {
		m.ExportFrom = e
	} else if n, ok := p.namedExports(); ok {
		m.NamedExports = n
	} else if v, ok := parserRecoverOk(p, p.variableStatement); ok {
		m.VariableStatement = v
	} else if d, ok := parserRecoverOk(p, p.declaration); ok {
		m.Declaration = d
	} else {
		panic("exportDeclaration: unimplemented")
	}
	return m
}

func (p *Parser) exportFrom() *ExportFrom {
	specifiers := p.exportFromClause()
	p.tokenizer.MustMatch(TFrom)
	moduleSpecifier := p.stringLiteral()
	p.automaticSemicolonInsertion()
	return &ExportFrom{
		ExportFromClause: specifiers,
		ModuleSpecifier:  moduleSpecifier,
	}
}

func (p *Parser) exportFromClause() (e *ExportFromClause) {
	if p.tokenizer.Match(TStar) {
		if p.tokenizer.Match(TAs) {
			e.StarAs, _ = p.moduleExportName()
			return
		} else {
			e.Star = true
			return
		}
	} else if n, ok := p.namedExports(); ok {
		e.NamedExports = n
		return
	}
	panic("exportFromClause: expected * or named exports")
}

func (p *Parser) moduleExportName() (m *ModuleExportName, ok bool) {
	m = &ModuleExportName{}
	if id, _ok := parserRecoverOk(p, p.bindingIdentifier); _ok {
		m.IdentifierName = id
		ok = _ok
		return
	} else if s, _ok := parserRecoverOk(p, p.stringLiteral); _ok {
		m.StringLiteral = s
		ok = _ok
		return
	} else {
		return
	}
}

func (p *Parser) namedExports() (n *NamedExports, ok bool) {
	specifiers, ok := p.exportSpecifierList()
	if !ok {
		return
	}
	return &NamedExports{
		ExportsList: specifiers,
	}, true
}

func (p *Parser) exportSpecifierList() (s *ExportsList, ok bool) {
	if !p.tokenizer.Match(TLeftBrace) {
		return
	}
	var items []*ExportSpecifier
	for {
		item, ok := p.exportSpecifier()
		if !ok {
			break
		}
		items = append(items, item)
		if p.tokenizer.Match(TComma) {
			continue
		}
		if p.tokenizer.CurrentToken.Type == TRightBrace {
			break
		}
	}
	p.tokenizer.MustMatch(TRightBrace)
	return &ExportsList{
		Items: items,
	}, true
}

func (p *Parser) exportSpecifier() (s *ExportSpecifier, ok bool) {
	name, ok := p.moduleExportName()
	if !ok {
		return
	}
	var alias *ModuleExportName
	if p.tokenizer.Match(TAs) {
		alias, ok = p.moduleExportName()
		if !ok {
			return
		}
	}
	return &ExportSpecifier{
		Name:  name,
		Alias: alias,
	}, true
}

func (p *Parser) importDeclaration() *ModuleItemImportDeclaration {
	p.tokenizer.MustMatch(TImport)
	importClause := p.importClause()
	p.tokenizer.MustMatch(TFrom)
	moduleSpecifier := p.stringLiteral()
	p.automaticSemicolonInsertion()
	return &ModuleItemImportDeclaration{
		ImportDeclaration: &ImportDeclaration{
			ImportClause:    importClause,
			ModuleSpecifier: moduleSpecifier,
		},
	}
}

func (p *Parser) importClause() *ImportClause {
	if identifier, ok := parserRecoverOk(p, p.bindingIdentifier); ok {
		return &ImportClause{
			ImportedDefaultBinding: identifier,
		}
	} else {
		return &ImportClause{
			NamedImports: p.importsList(),
		}
	}
}

func (p *Parser) importsList() *ImportsList {
	p.tokenizer.MustMatch(TLeftBrace)
	var items []*ImportSpecifier
	for {
		item, ok := p.importSpecifier()
		if !ok {
			break
		}
		items = append(items, item)
		if p.tokenizer.Match(TComma) {
			continue
		}
		if p.tokenizer.CurrentToken.Type == TRightBrace {
			break
		}
	}
	p.tokenizer.MustMatch(TRightBrace)
	return &ImportsList{
		Items: items,
	}
}

func (p *Parser) importSpecifier() (s *ImportSpecifier, ok bool) {
	name, ok := p.moduleExportName()
	if !ok {
		return
	}
	if p.tokenizer.Match(TAs) {
		alias := p.bindingIdentifier()
		return &ImportSpecifier{
			ModuleExportName: name,
			ImportedBinding:  alias,
		}, true
	}
	return &ImportSpecifier{
		ImportedBinding: name.IdentifierName,
	}, true
}

// MARK: - ParserContext

type ParserContext struct {
	FileName string
	BaseDir  string
}

// MARK: - Precedence

type (
	precedence                 int
	precedenceAssociativityAlt int
)

const (
	precAssocNewArgs precedenceAssociativityAlt = iota
	precAssocFunctionArgs
	precAssocPostfixIncrement
	precAssocPostfixDecrement
	precAssocUnaryPlus
	precAssocUnaryMinus
	precAssocPrefixIncrement
	precAssocPrefixDecrement
)

type associativity int

const (
	associativeNone associativity = iota
	associativeLeft
	associativeRight
)

type acceptContext struct {
	precedence    precedence
	associativity associativity
}

func (p *Parser) acceptContextLowest() *acceptContext {
	return &acceptContext{
		precedence: 0,
	}
}

// add 1 to the precedence
func (p *Parser) acceptContextHigherThan(t TokenType) *acceptContext {
	c := p.acceptContext(t)
	c.precedence += 1
	return c
}

func (p *Parser) acceptContext(t TokenType) *acceptContext {
	switch t {
	case TLeftParen:
		return &acceptContext{
			precedence: 18,
		}
	case TDot, TQuestionDot:
		return &acceptContext{
			precedence:    17,
			associativity: associativeLeft,
		}
	case TLeftBracket:
		return &acceptContext{
			precedence: 17,
		}
	case TNew:
		return &acceptContext{
			precedence: 16,
		}
	case TPlusPlus, TMinusMinus:
		return &acceptContext{
			precedence: 15,
		}
	case TWave, TTypeof, TDelete, TNot:
		return &acceptContext{
			precedence: 14,
		}
	case TStarStar:
		return &acceptContext{
			precedence:    13,
			associativity: associativeRight,
		}
	case TStar, TSlash, TPercent:
		return &acceptContext{
			precedence:    12,
			associativity: associativeLeft,
		}
	case TPlus, TMinus:
		return &acceptContext{
			precedence:    11,
			associativity: associativeLeft,
		}
	case TLeftShift, TRightShift, TUnsignedRightShift:
		return &acceptContext{
			precedence:    10,
			associativity: associativeLeft,
		}
	case TLessThan, TLessThanEquals, TGreaterThan, TGreaterThanEquals, TInstanceof, TIn:
		return &acceptContext{
			precedence:    9,
			associativity: associativeLeft,
		}
	case TEqualsEquals, TNotEquals, TStrictEquals, TStrictNotEquals:
		return &acceptContext{
			precedence:    8,
			associativity: associativeLeft,
		}
	case TAmpersand:
		return &acceptContext{
			precedence:    7,
			associativity: associativeLeft,
		}
	case TCaret:
		return &acceptContext{
			precedence:    6,
			associativity: associativeLeft,
		}
	case TPipe:
		return &acceptContext{
			precedence:    5,
			associativity: associativeLeft,
		}
	case TAmpersandAmpersand:
		return &acceptContext{
			precedence:    4,
			associativity: associativeLeft,
		}
	case TPipePipe, TQuestionQuestion:
		return &acceptContext{
			precedence:    3,
			associativity: associativeLeft,
		}
	case TEquals, TPlusEquals, TMinusEquals, TStarEquals, TStarStarEquals, TPercentEquals, TLeftShiftEquals, TRightShiftEquals, TUnsignedRightShiftEquals, TAmpersandEquals, TCaretEquals, TPipeEquals:
		return &acceptContext{
			precedence:    2,
			associativity: associativeRight,
		}
	case TQuestion:
		return &acceptContext{
			precedence:    2,
			associativity: associativeRight,
		}
	case TArrow:
		return &acceptContext{
			precedence:    2,
			associativity: associativeRight,
		}
	case TYield, TDotDotDot:
		return &acceptContext{
			precedence: 2,
		}
	case TComma:
		return &acceptContext{
			precedence:    1,
			associativity: associativeLeft,
		}
	default:
		return p.acceptContextLowest()
	}
}

func (p *Parser) precedence(t TokenType) precedence {
	return p.acceptContext(t).precedence
}

func (p *Parser) acceptContextAlt(flag precedenceAssociativityAlt) *acceptContext {
	switch flag {
	case precAssocNewArgs, precAssocFunctionArgs:
		return &acceptContext{
			precedence: 17,
		}
	case precAssocPostfixIncrement, precAssocPostfixDecrement:
		return &acceptContext{
			precedence: 15,
		}
	case precAssocPrefixIncrement, precAssocPrefixDecrement, precAssocUnaryPlus, precAssocUnaryMinus:
		return &acceptContext{
			precedence: 14,
		}
	}
	panic("unreachable")
}

func (p *Parser) precedenceAlt(flag precedenceAssociativityAlt) precedence {
	return p.acceptContextAlt(flag).precedence
}

func (p *Parser) associativity(t TokenType) associativity {
	return p.acceptContext(t).associativity
}

func (p *Parser) associativityAlt(flag precedenceAssociativityAlt) associativity {
	return p.acceptContextAlt(flag).associativity
}

// MARK: - Parse

func (p *Parser) ParseNode() StatementList {
	return p.statementList()
}

func (p *Parser) statementList() (list StatementList) {
	for {
		item := p.statementListItem()
		if item == nil {
			break
		}
		switch stmt := item.(type) {
		case *StatementListItemStatement:
			if _, ok := stmt.Statement.(*StatementEmpty); ok {
				break
			}
		}
		list = append(list, item)
	}

	return
}

func (p *Parser) functionBody(functionType FunctionType) *FunctionBody {
	inFunctionBodyBefore := p.inFunctionBody
	p.inFunctionBody = true
	defer func() {
		p.inFunctionBody = inFunctionBodyBefore
	}()

	list := p.statementList()
	return &FunctionBody{
		StatementList: list,
		Type:          functionType,
	}
}

func (p *Parser) statementListItem() (stmt StatementListItem) {
	t := p.tokenizer.CurrentToken
	switch t.Type {
	case TAsync, TFunction, TLet, TConst, TClass:
		d := p.declaration()
		stmt = &StatementListItemDeclaration{
			Declaration: d,
		}
	default:
		s := p.statement()
		if s == nil {
			p.automaticSemicolonInsertion()
			return nil
		}

		stmt = &StatementListItemStatement{
			Statement: s,
		}
	}
	return stmt
}

func (p *Parser) automaticSemicolonInsertion() {
	p.tokenizer.Match(TSemicolon)
}

func (p *Parser) statement() Statement {
	t := p.tokenizer.CurrentToken
	switch t.Type {
	case TVar:
		return p.variableStatement()
	case TSemicolon:
		p.tokenizer.Next()
		return &StatementEmpty{}
	case TEOF:
		return nil
	case TLeftBrace:
		return p.blockStatement()
	case TDebugger:
		p.tokenizer.Next()
		return &StatementDebugger{}
	case TIf:
		return p.ifStatement()
	case TWhile, TDo, TFor:
		return p.breakableStatement()
	case TThrow:
		return p.throwStatement()
	case TTry:
		return p.tryStatement()
	case TRightBrace:
		return nil
	case TReturn:
		return p.returnStatement()
	case TContinue:
		return p.continueStatement()
	case TBreak:
		return p.breakStatement()
	default:
		return p.expressionStatement()
	}
}

func (p *Parser) breakStatement() *StatementBreak {
	p.tokenizer.MustMatch(TBreak)
	if !p.inBreakable {
		panic("breakStatement: not in breakable")
	}

	t := p.tokenizer.CurrentToken
	var label string
	if t.Type == TIdentifier {
		label = t.Value
		p.tokenizer.Next()
	}
	p.automaticSemicolonInsertion()
	return &StatementBreak{
		Label: IdentifierName(label),
	}
}

func (p *Parser) continueStatement() *StatementContinue {
	p.tokenizer.MustMatch(TContinue)
	if !p.inIteration {
		panic("continueStatement: not in iteration")
	}
	t := p.tokenizer.CurrentToken
	var label string
	if t.Type == TIdentifier {
		label = t.Value
		p.tokenizer.Next()
	}
	p.automaticSemicolonInsertion()
	return &StatementContinue{
		Label: IdentifierName(label),
	}
}

func (p *Parser) throwStatement() *StatementThrow {
	p.tokenizer.MustMatch(TThrow)
	p.noLineTerminatorHere()
	expr := p.expression(p.acceptContextLowest())
	return &StatementThrow{
		Expression: expr,
	}
}

func (p *Parser) tryStatement() *StatementTry {
	p.tokenizer.MustMatch(TTry)
	block := p.block()
	var catch *Block
	var finally *Block
	var catchParameter CatchParameter
	if p.tokenizer.CurrentToken.Type == TCatch {
		p.tokenizer.Next()
		if p.tokenizer.CurrentToken.Type == TLeftParen {
			p.tokenizer.Next()
			catchParameter = CatchParameter(p.bindingIdentifier())
			p.tokenizer.MustMatch(TRightParen)
		}
		catch = p.block()
	}
	if p.tokenizer.CurrentToken.Type == TFinally {
		finally = p.block()
	}
	if catch == nil && finally == nil {
		panic("tryStatement: expected catch or finally")
	}
	return &StatementTry{
		CatchParameter: catchParameter,
		TryBlock:       block,
		CatchBlock:     catch,
		FinallyBlock:   finally,
	}
}

func (p *Parser) bindingRestElement() (b BindingRestElement, ok bool) {
	if !p.tokenizer.Match(TDotDotDot) {
		return
	}
	identifier := p.bindingIdentifier()
	b = &BindingRestElementIdentifier{
		Identifier: identifier,
	}
	return b, true
}

func (p *Parser) bindingElement() (b *BindingElement, ok bool) {
	identifier := p.bindingIdentifier()
	var init Expression
	if p.tokenizer.Match(TEquals) {
		init = p.expression(p.acceptContextLowest())
	}
	b = &BindingElement{
		SingleNameBinding: &SingleNameBinding{
			BindingIdentifier: identifier,
			Initializer:       init,
		},
	}
	return b, true
}

func (p *Parser) formalParameters() *FormalParameters {
	var items []FormalParametersItem
	for {
		t := p.tokenizer.CurrentToken
		if t.Type == TRightParen {
			break
		}
		if b, ok := p.bindingRestElement(); ok {
			items = append(items, &FormalParameterFunctionRestParameter{
				BindingRestElement: b,
			})
			p.tokenizer.Match(TComma)
			break
		} else if b, ok := p.bindingElement(); ok {
			items = append(items, &FormalParameter{
				BindingElement: b,
			})
			p.tokenizer.Match(TComma)
		} else {
			panic("formalParameters: expected binding element")
		}
	}

	return &FormalParameters{
		Items: items,
	}
}

func (p *Parser) functionDeclaration() *FunctionDeclaration {
	startOffset := p.tokenizer.CurrentStartIndex()
	p.tokenizer.MustMatch(TFunction)
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	functionBody := p.functionBody(FunctionTypeNormal)
	endIndex := p.tokenizer.CurrentEndIndex()
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:endIndex]
	return &FunctionDeclaration{
		Identifier:       identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             functionBody,
	}
}

func (p *Parser) asyncArrowFunction() *PrimaryExpressionAsyncArrowFunction {
	startOffset := p.tokenizer.Index
	p.tokenizer.Match(TAsync)
	var params *FormalParameters
	if p.tokenizer.Match(TLeftParen) {
		params = p.formalParameters()
		p.tokenizer.MustMatch(TRightParen)
	} else {
		identifier := p.bindingIdentifier()
		params = &FormalParameters{
			Items: []FormalParametersItem{
				&FormalParameter{
					BindingElement: &BindingElement{
						SingleNameBinding: &SingleNameBinding{
							BindingIdentifier: identifier,
						},
					},
				},
			},
		}
	}
	p.tokenizer.MustMatch(TArrow)
	var body *FunctionBody
	if p.tokenizer.Match(TLeftBrace) {
		body = p.functionBody(FunctionTypeAsync)
		p.tokenizer.MustMatch(TRightBrace)
	} else {
		e := p.expression(p.acceptContextHigherThan(TComma))
		body = &FunctionBody{
			StatementList: StatementList{
				&StatementListItemStatement{
					Statement: &StatementExpression{
						Expression: e,
					},
				},
			},
		}
	}
	sourceText := p.SourceText[startOffset:p.tokenizer.Index]
	return &PrimaryExpressionAsyncArrowFunction{
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             body,
	}
}

func (p *Parser) asyncFunctionExpression() *PrimaryExpressionAsyncFunctionExpression {
	startOffset := p.tokenizer.Index
	p.tokenizer.Match(TAsync)
	p.tokenizer.MustMatch(TFunction)
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	body := p.functionBody(FunctionTypeAsync)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.Index]
	return &PrimaryExpressionAsyncFunctionExpression{
		Identifier:       identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             body,
	}
}

func (p *Parser) asyncGeneratorExpression() *PrimaryExpressionAsyncGeneratorExpression {
	startOffset := p.tokenizer.CurrentStartIndex()
	p.tokenizer.Match(TAsync)
	p.tokenizer.MustMatch(TFunction)
	p.tokenizer.MustMatch(TStar)
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	body := p.functionBody(FunctionTypeAsyncGenerator)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.CurrentStartIndex()]
	return &PrimaryExpressionAsyncGeneratorExpression{
		IdentifierName:   identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             body,
	}
}

func (p *Parser) generatorExpression() *PrimaryExpressionGeneratorExpression {
	startOffset := p.tokenizer.CurrentStartIndex()
	p.tokenizer.MustMatch(TFunction)
	p.tokenizer.MustMatch(TStar)
	var identifier IdentifierName
	if id, ok := parserRecoverOk(p, p.bindingIdentifier); ok {
		identifier = id
	}
	p.tokenizer.MustMatch(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	body := p.functionBody(FunctionTypeGenerator)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.CurrentStartIndex()]
	return &PrimaryExpressionGeneratorExpression{
		IdentifierName:   identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             body,
	}
}

func (p *Parser) functionExpression() *PrimaryExpressionFunctionExpression {
	startOffset := p.tokenizer.CurrentStartIndex()
	p.tokenizer.MustMatch(TFunction)
	var identifier IdentifierName
	if p.tokenizer.CurrentToken.Type == TIdentifier {
		identifier = p.bindingIdentifier()
	}
	p.tokenizer.MustMatch(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	functionBody := p.functionBody(FunctionTypeNormal)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.CurrentStartIndex()]
	return &PrimaryExpressionFunctionExpression{
		Identifier:       identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             functionBody,
	}
}

func (p *Parser) noLineTerminatorHere() {
	// TODO
}

func (p *Parser) classDeclaration() *ClassDeclaration {
	startIndex := p.tokenizer.CurrentStartIndex()
	p.tokenizer.MustMatch(TClass)
	identifier := p.bindingIdentifier()
	classTail := p.classTail()
	sourceText := p.SourceText[startIndex:p.tokenizer.CurrentStartIndex()]
	return &ClassDeclaration{
		IdentifierName: identifier,
		ClassTail:      classTail,
		SourceText:     sourceText,
	}
}

func (p *Parser) classTail() *ClassTail {
	var classHeritage Expression
	if p.tokenizer.Match(TExtends) {
		classHeritage = p.expression(p.acceptContextLowest())
	}
	p.tokenizer.MustMatch(TLeftBrace)
	classBody := p.classBody()
	p.tokenizer.MustMatch(TRightBrace)
	return &ClassTail{
		ClassHeritage: classHeritage,
		ClassBody:     classBody,
	}
}

func (p *Parser) classBody() *ClassBody {
	var items []ClassElement
	inClassBody := p.inClassBody
	p.inClassBody = true
	defer func() {
		p.inClassBody = inClassBody
	}()
	for {
		if p.tokenizer.CurrentToken.Type == TRightBrace {
			break
		}
		items = append(items, p.classElement())
	}
	return &ClassBody{
		ClassElementList: &ClassElementList{
			Items: items,
		},
	}
}

func (p *Parser) classElement() ClassElement {
	defer func() {
		p.tokenizer.Match(TSemicolon)
	}()
	if p.tokenizer.Match(TStatic) {
		if p.tokenizer.Match(TLeftBrace) {
			statementList := p.statementList()
			p.tokenizer.MustMatch(TRightBrace)
			return &ClassElementStaticBlock{
				StatementList: statementList,
			}
		}
		if def, ok := parserRecoverOk(p, func() *MethodDefinition {
			return p.methodDefinition(MethodDefinitionTypeNil)
		}); ok {
			return &ClassElementStaticMethodDefinition{
				MethodDefinition: def,
			}
		} else {
			field := p.fieldDefinition()
			return &ClassElementStaticFieldDefinition{
				FieldDefinition: field,
			}
		}
	} else {
		if def, ok := parserRecoverOk(p, func() *MethodDefinition {
			return p.methodDefinition(MethodDefinitionTypeNil)
		}); ok {
			return &ClassElementMethodDefinition{
				MethodDefinition: def,
			}
		} else {
			field := p.fieldDefinition()
			return &ClassElementFieldDefinition{
				FieldDefinition: field,
			}
		}
	}
	return &ClassElementEmpty{}
}

func (p *Parser) fieldDefinition() *FieldDefinition {
	propertyName, ok := p.propertyName()
	Assert(ok)
	var initializer Expression
	if p.tokenizer.Match(TEquals) {
		initializer = p.expression(p.acceptContextLowest())
	}
	return &FieldDefinition{
		PropertyName: propertyName,
		Initializer:  initializer,
	}
}

func (p *Parser) lexicalDeclaration() *LexicalDeclaration {
	t := p.tokenizer.CurrentToken
	var lexicalType LetOrConst
	if t.Type == TLet {
		lexicalType = LetOrConstLet
	} else if t.Type == TConst {
		lexicalType = LetOrConstConst
	} else {
		panic("lexicalDeclaration: expected let or const")
	}
	p.tokenizer.Next()
	list := p.bindingList()
	p.automaticSemicolonInsertion()
	return &LexicalDeclaration{
		Type:        lexicalType,
		BindingList: list,
	}
}

func (p *Parser) bindingList() *BindingList {
	var items []*LexicalBinding
	for {
		item := p.lexicalBinding()
		items = append(items, item)
		if p.tokenizer.CurrentToken.Type == TComma {
			p.tokenizer.Next()
		} else {
			break
		}
	}
	return &BindingList{
		Items: items,
	}
}

func (p *Parser) lexicalBinding() *LexicalBinding {
	identifier := p.bindingIdentifier()
	var init Expression
	if p.tokenizer.CurrentToken.Type == TEquals {
		p.tokenizer.Next()
		init = p.expression(p.acceptContextHigherThan(TComma))
	}
	return &LexicalBinding{
		Identifier:  identifier,
		Initializer: init,
	}
}

func (p *Parser) hoistableDeclaration() DeclarationHoistable {
	t := p.tokenizer.CurrentToken
	if t.Type == TFunction && p.tokenizer.NextToken.Type == TStar {
		d := p.generatorDeclaration()
		return &DeclarationHoistableGenerator{
			GeneratorDeclaration: d,
		}
	} else if t.Type == TFunction {
		functionDeclaration := p.functionDeclaration()
		return &DeclarationHoistableFunction{
			FunctionDeclaration: functionDeclaration,
		}
	} else if t.Type == TAsync {
		startOffset := p.tokenizer.CurrentStartIndex()
		p.tokenizer.MustMatch(TAsync)
		p.tokenizer.MustMatch(TFunction)
		if p.tokenizer.CurrentToken.Type == TStar {
			d := p.asyncGeneratorDeclaration(startOffset)
			return &DeclarationHoistableAsyncGenerator{
				AsyncGeneratorDeclaration: d,
			}
		}
		return &DeclarationHoistableAsyncFunction{
			AsyncFunctionDeclaration: p.asyncFunctionDeclaration(startOffset),
		}

	}
	panic("unimplemented")
}

func (p *Parser) asyncFunctionDeclaration(startOffset int) *AsyncFunctionDeclaration {
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	body := p.functionBody(FunctionTypeAsync)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.CurrentStartIndex()]
	return &AsyncFunctionDeclaration{
		Identifier:       identifier,
		FormalParameters: params,
		SourceText:       sourceText,
		Body:             body,
	}
}

func (p *Parser) asyncGeneratorDeclaration(startOffset int) *AsyncGeneratorDeclaration {
	p.tokenizer.MustMatch(TStar)
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	formalParams := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	body := p.functionBody(FunctionTypeAsyncGenerator)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.CurrentStartIndex()]
	return &AsyncGeneratorDeclaration{
		Identifier:       identifier,
		FormalParameters: formalParams,
		SourceText:       sourceText,
		Body:             body,
	}
}

func (p *Parser) generatorDeclaration() *GeneratorDeclaration {
	startOffset := p.tokenizer.CurrentStartIndex()
	p.tokenizer.MustMatch(TFunction)
	p.tokenizer.MustMatch(TStar)
	identifier := p.bindingIdentifier()
	p.tokenizer.MustMatch(TLeftParen)
	formalParams := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	functionBody := p.functionBody(FunctionTypeGenerator)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[startOffset:p.tokenizer.CurrentStartIndex()]
	return &GeneratorDeclaration{
		Identifier:       identifier,
		FormalParameters: formalParams,
		SourceText:       sourceText,
		Body:             functionBody,
	}
}

func (p *Parser) declaration() Declaration {
	t := p.tokenizer.CurrentToken
	if t.Type == TFunction || t.Type == TAsync {
		return p.hoistableDeclaration()
	}
	if t.Type == TClass {
		return p.classDeclaration()
	}
	if t.Type == TLet || t.Type == TConst {
		return p.lexicalDeclaration()
	}
	panic("unimplemented")
}

func (p *Parser) breakableStatement() *BreakableStatement {
	inBreakable := p.inBreakable
	p.inBreakable = true
	defer func() {
		p.inBreakable = inBreakable
	}()
	return &BreakableStatement{
		IterationStatement: p.iterationStatement(),
	}
}

// MARK: - Iteration

func (p *Parser) iterationStatement() IterationStatement {
	t := p.tokenizer.CurrentToken
	inIteration := p.inIteration
	p.inIteration = true
	defer func() {
		p.inIteration = inIteration
	}()

	switch t.Type {
	case TDo:
		return p.doWhileStatement()
	case TWhile:
		return p.whileStatement()
	case TFor:
		if e, ok := parserRecoverOk(p, p.forInOfStatement); ok {
			return e
		}
		return p.forStatement()
	default:
		panic("iterationStatement: expected do, while or for")
	}
}

func (p *Parser) forInOfStatement() *ForInOfStatement {
	p.tokenizer.MustMatch(TFor)
	var isAwait bool
	if p.tokenizer.Match(TAwait) {
		isAwait = true
	}
	p.tokenizer.MustMatch(TLeftParen)
	init := &ForInOfStatementInitializer{}
	if p.tokenizer.Match(TVar) {
		init.ForBinding = p.forBinding()
	} else if p.tokenizer.CurrentToken.Type == TLet || p.tokenizer.CurrentToken.Type == TConst {
		init.ForDeclaration = p.forDeclaration()
	} else {
		init.LeftHandSideExpression = p.expression(p.acceptContextLowest())
	}

	var statementType ForInOfStatementType
	if p.tokenizer.Match(TIn) {
		statementType = ForInOfStatementTypeIn
	} else if p.tokenizer.Match(TOf) {
		statementType = ForInOfStatementTypeOf
	} else {
		panic("forInOfStatement: expected in or of")
	}

	expression := p.expression(p.acceptContextLowest())
	p.tokenizer.MustMatch(TRightParen)
	body := p.statement()

	return &ForInOfStatement{
		Type:        statementType,
		IsAwait:     isAwait,
		Initializer: init,
		Expression:  expression,
		Body:        body,
	}
}

func (p *Parser) forDeclaration() *ForDeclaration {
	var letOrConst LetOrConst
	if p.tokenizer.Match(TLet) {
		letOrConst = LetOrConstLet
	} else if p.tokenizer.Match(TConst) {
		letOrConst = LetOrConstConst
	} else {
		panic("forDeclaration: expected let or const")
	}
	return &ForDeclaration{
		LetOrConst: letOrConst,
		ForBinding: p.forBinding(),
	}
}

func (p *Parser) forBinding() *ForBinding {
	f := &ForBinding{}
	if identifier, ok := parserRecoverOk(p, p.bindingIdentifier); ok {
		f.BindingIdentifier = identifier
	} else {
		f.BindingPattern = p.bindingPattern()
	}

	return f
}

func (p *Parser) bindingPattern() *BindingPattern {
	b := &BindingPattern{}
	// TODO:
	panic("unimplemented")
	return b
}

func (p *Parser) forStatement() *ForStatement {
	p.tokenizer.MustMatch(TFor)
	p.tokenizer.MustMatch(TLeftParen)
	var init ForStatementInitializer
	t := p.tokenizer.CurrentToken
	if t.Type == TVar {
		init = &ForStatementInitializerVariable{
			VariableStatement: p.variableStatement(),
		}
	} else if t.Type == TLet || t.Type == TConst {
		init = &ForStatementInitializerLexicalDeclaration{
			LexicalDeclaration: p.lexicalDeclaration(),
		}
	} else {
		init = &ForStatementInitializerExpression{
			Expression: p.expression(p.acceptContextLowest()),
		}
	}
	// semicolon has been consumed by initializer
	// p.tokenizer.Match(TSemicolon)
	var condition Expression
	if p.tokenizer.CurrentToken.Type != TSemicolon {
		condition = p.expression(p.acceptContextLowest())
	}
	p.tokenizer.MustMatch(TSemicolon)
	var increment Expression
	if p.tokenizer.CurrentToken.Type != TRightParen {
		increment = p.expression(p.acceptContextLowest())
	}
	p.tokenizer.MustMatch(TRightParen)
	body := p.statement()

	return &ForStatement{
		Initializer: init,
		Condition:   condition,
		Increment:   increment,
		Body:        body,
	}
}

func (p *Parser) doWhileStatement() *StatementDoWhile {
	p.tokenizer.MustMatch(TDo)
	body := p.statement()
	p.tokenizer.MustMatch(TWhile)
	p.tokenizer.MustMatch(TLeftParen)
	condition := p.expression(p.acceptContextLowest())
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TSemicolon)

	return &StatementDoWhile{
		Body:      body,
		Condition: condition,
	}
}

func (p *Parser) whileStatement() *StatementWhile {
	p.tokenizer.MustMatch(TWhile)
	p.tokenizer.MustMatch(TLeftParen)
	condition := p.expression(p.acceptContextLowest())
	p.tokenizer.MustMatch(TRightParen)
	body := p.statement()

	return &StatementWhile{
		Condition: condition,
		Body:      body,
	}
}

// MARK: - CompletionTypeReturn

func (p *Parser) returnStatement() *StatementReturn {
	p.tokenizer.MustMatch(TReturn)
	t := p.tokenizer.CurrentToken

	if !p.inFunctionBody {
		panic("returnStatement: not in function")
	}

	if t.Type == TSemicolon {
		p.tokenizer.Next()
		return &StatementReturn{}
	}
	expr := p.expression(p.acceptContextLowest())
	p.automaticSemicolonInsertion()
	return &StatementReturn{
		Expression: expr,
	}
}

// MARK: - Condition

func (p *Parser) ifStatement() *IfStatement {
	p.tokenizer.MustMatch(TIf)
	p.tokenizer.MustMatch(TLeftParen)
	condition := p.expression(p.acceptContextLowest())
	p.tokenizer.MustMatch(TRightParen)
	consequent := p.statement()

	var alternate Statement
	if p.tokenizer.Match(TElse) {
		alternate = p.statement()
	}

	return &IfStatement{
		Condition:  condition,
		Consequent: consequent,
		Alternate:  alternate,
	}
}

func (p *Parser) expressionStatement() *StatementExpression {
	expr := p.expression(p.acceptContextLowest())
	p.automaticSemicolonInsertion()
	if expr == nil {
		panic("expressionStatement: expected expression")
	}

	return &StatementExpression{
		Expression: expr,
	}
}

func (p *Parser) importCall() (*ExpressionImportCall, bool) {
	t := p.tokenizer.CurrentToken
	if t.Type != TImport {
		return nil, false
	}
	p.tokenizer.Next()
	p.tokenizer.MustMatch(TLeftParen)
	e := p.expression(p.acceptContextLowest())
	p.tokenizer.MustMatch(TRightParen)
	p.automaticSemicolonInsertion()
	return &ExpressionImportCall{
		Expression: e,
	}, true
}

func (p *Parser) superCall() (*ExpressionSuperCall, bool) {
	t := p.tokenizer.CurrentToken
	if t.Type != TSuper {
		return nil, false
	}
	p.tokenizer.Next()
	args := p.arguments()
	if !p.inClassConstructor {
		panic("superCall: not in class constructor")
	}
	return &ExpressionSuperCall{
		Arguments: args,
	}, true
}

func (p *Parser) followedByLineTerminator(previous Token) bool {
	gap := p.tokenizer.SourceText[previous.EndIndex:p.tokenizer.CurrentToken.StartIndex]
	// TODO: check in `lineTerminators`
	return strings.Contains(string(gap), "\n")
}

func (p *Parser) yieldExpression() (*YieldExpression, bool) {
	t := p.tokenizer.CurrentToken
	if !p.tokenizer.Match(TYield) {
		return nil, false
	}
	if p.followedByLineTerminator(t) {
		return &YieldExpression{AssignmentExpression: nil}, true
	}
	e := p.expression(p.acceptContextHigherThan(TYield))
	return &YieldExpression{
		AssignmentExpression: e,
	}, true
}

func (p *Parser) newExpression() (*NewExpression, bool) {
	t := p.tokenizer.CurrentToken

	previous := p.callExpressionForbidden
	p.callExpressionForbidden = true
	defer func() {
		p.callExpressionForbidden = previous
	}()

	if t.Type != TNew {
		return nil, false
	}
	p.tokenizer.Next()
	accept := p.acceptContext(TNew)
	expr := p.expression(accept)
	p.callExpressionForbidden = previous
	// can be called without parens, like: `new Date`
	var args Arguments
	if p.tokenizer.CurrentToken.Type == TLeftParen {
		args = p.arguments()
	}
	p.automaticSemicolonInsertion()
	return &NewExpression{
		Callee:    expr,
		Arguments: args,
	}, true
}

// tryUnaryExpression accept unary token
// if token is not unary, return nil
func (p *Parser) tryUnaryExpression() (Expression, bool) {
	t := p.tokenizer.CurrentToken
	var operator UnaryOperator
	unaryMap := map[TokenType]UnaryOperator{
		TDelete: UnaryOperatorDelete,
		TVoid:   UnaryOperatorVoid,
		TTypeof: UnaryOperatorTypeof,
		TPlus:   UnaryOperatorAddition,
		TMinus:  UnaryOperatorSubtraction,
		TNot:    UnaryOperatorLogicalNot,
		TTilde:  UnaryOperatorBitwiseNot,
	}
	if op, ok := unaryMap[t.Type]; ok {
		p.tokenizer.Next()
		operator = op
	} else {
		return nil, false
	}
	var accept *acceptContext
	if t.Type == TPlus {
		accept = p.acceptContextAlt(precAssocUnaryPlus)
	} else if t.Type == TMinus {
		accept = p.acceptContextAlt(precAssocUnaryMinus)
	} else {
		accept = p.acceptContext(t.Type)
	}
	expr := p.expression(accept)
	return &UnaryExpression{
		Operator: operator,
		Operand:  expr,
	}, true
}

func (p *Parser) metaProperty() (MetaProperty, bool) {
	newTarget, ok := p.newTarget()
	if ok {
		return newTarget, true
	}
	return nil, false
}

func (p *Parser) importMeta() *MetaPropertyImportMeta {
	p.tokenizer.MustMatch(TImport)
	p.tokenizer.MustMatch(TDot)
	p.tokenizer.MustMatch(TIdentifier)
	if !p.inModule {
		panic("importMeta: not in module")
	}
	return &MetaPropertyImportMeta{}
}

func (p *Parser) newTarget() (m *MetaPropertyNewTarget, ok bool) {
	t := p.tokenizer.CurrentToken
	if t.Type != TNew || p.tokenizer.NextToken.Type != TDot {
		return
	}
	p.tokenizer.MustMatch(TNew)
	p.tokenizer.MustMatch(TDot)
	identifier := p.tokenizer.CurrentToken
	if identifier.Value != "target" {
		return
	}
	return &MetaPropertyNewTarget{}, true
}

func (p *Parser) updateExpression(primaryExpression Expression) (*UpdateExpression, bool) {
	t := p.tokenizer.CurrentToken
	var operator UpdateOperator
	if op, ok := UpdateOperatorMap[t.Type]; ok {
		operator = op
	} else {
		return nil, false
	}
	p.tokenizer.Next()
	var expr Expression
	var updateType UpdateExpressionType
	if primaryExpression == nil {
		expr = p.expression(p.acceptContextAlt(precAssocPrefixIncrement))
		updateType = UpdateExpressionTypePrefix
	} else {
		expr = primaryExpression
		updateType = UpdateExpressionTypePostfix
	}

	if updateType == UpdateExpressionTypePrefix && expr.AssignmentTargetType() != AssignmentTargetTypeSimple {
		panic("updateExpression: invalid assignment target for prefix")
	}
	if updateType == UpdateExpressionTypePrefix && expr.AssignmentTargetType() != AssignmentTargetTypeSimple {
		panic("updateExpression: invalid assignment target for postfix")
	}

	return &UpdateExpression{
		Operator: operator,
		Type:     updateType,
		Operand:  expr,
	}, true
}

func (p *Parser) expression(accept *acceptContext) Expression {
	var expr Expression
	if unary, ok := p.tryUnaryExpression(); ok {
		expr = unary
	} else if meta, ok := p.metaProperty(); ok {
		expr = meta
	} else if update, ok := p.updateExpression(nil); ok {
		expr = update
	} else if super, ok := p.superProperty(); ok {
		expr = super
	} else if super, ok := p.superCall(); ok {
		expr = super
	} else if i, ok := p.importCall(); ok {
		expr = i
	} else if newExpression, ok := p.newExpression(); ok {
		expr = newExpression
	} else if yieldExpression, ok := p.yieldExpression(); ok {
		expr = yieldExpression
	} else {
		expr = p.primaryExpression()
	}

	for {
		nextToken := p.tokenizer.CurrentToken
		newAcceptContext := p.acceptContext(nextToken.Type)
		if newAcceptContext.precedence < accept.precedence {
			return expr
		}
		if newAcceptContext.precedence == accept.precedence && newAcceptContext.associativity == associativeNone {
			return expr
		}
		if newAcceptContext.precedence == accept.precedence && newAcceptContext.associativity == associativeLeft {
			return expr
		}

		secondary := p.secondaryExpression(expr, newAcceptContext)
		if secondary == nil {
			return expr
		}
		expr = secondary
	}
}

func (p *Parser) secondaryExpression(left Expression, accept *acceptContext) Expression {
	t := p.tokenizer.CurrentToken
	switch t.Type {
	case TQuestionDot:
		return p.optionalExpression(left)
	case TLeftParen:
		if p.callExpressionForbidden {
			return nil
		}
		return p.callExpression(left)
	case TLeftBracket, TDot:
		return p.memberExpression(left)
	case TPlusPlus, TMinusMinus:
		update, ok := p.updateExpression(left)
		if !ok {
			panic("secondaryExpression: expected update expression")
		}
		return update
	case TLessThan,
		TLessThanEquals,
		TGreaterThan,
		TGreaterThanEquals,
		TInstanceof,
		TIn:
		return p.relationalExpression(left, accept)
	case TEqualsEquals,
		TNotEquals,
		TStrictEquals,
		TStrictNotEquals:
		return p.equalityExpression(left, accept)
	case TAmpersandAmpersand, TPipePipe, TQuestionQuestion:
		return p.logicalExpression(left, accept)
	case TQuestion:
		return p.conditionalExpression(left, accept)
	case TComma:
		return p.sequenceExpression(left)
	case TStar,
		TStarStar,
		TSlash,
		TPercent,
		TPlus,
		TMinus,
		TLeftShift,
		TRightShift,
		TUnsignedRightShift,
		TAmpersand,
		TCaret,
		TPipe:
		return p.binaryExpression(left, accept)
	case TEquals,
		TPlusEquals,
		TMinusEquals,
		TStarEquals,
		TStarStarEquals,
		TPercentEquals,
		TLeftShiftEquals,
		TRightShiftEquals,
		TUnsignedRightShiftEquals,
		TAmpersandEquals,
		TCaretEquals,
		TPipeEquals,
		TAmpersandAmpersandEquals,
		TPipePipeEquals,
		TQuestionQuestionEquals:
		return p.assignmentExpression(left, accept)
	default:
		panic("secondaryExpression: unexpected token")
	}
	return left
}

func (p *Parser) optionalExpression(left Expression) *OptionalExpression {
	p.tokenizer.MustMatch(TQuestionDot)
	var arguments Arguments
	var expr Expression
	var identifier IdentifierName
	if p.tokenizer.Match(TLeftParen) {
		arguments = p.arguments()
	} else if p.tokenizer.Match(TLeftBracket) {
		expr = p.expression(p.acceptContextLowest())
		p.tokenizer.MustMatch(TRightBracket)
	} else {
		identifier = p.identifierReference().Identifier
	}
	return &OptionalExpression{
		Expr: left,
		Property: &OptionalExpressionProperty{
			Arguments:  arguments,
			Expression: expr,
			Identifier: identifier,
		},
	}
}

func (p *Parser) assignmentExpression(left Expression, accept *acceptContext) *AssignmentExpression {
	t := p.tokenizer.CurrentToken
	p.tokenizer.Next()
	right := p.expression(accept)
	return &AssignmentExpression{
		Operator: operatorAssignmentMap[t.Type],
		Left:     left,
		Right:    right,
	}
}

func (p *Parser) binaryExpression(left Expression, accept *acceptContext) *ExpressionBinaryExpression {
	t := p.tokenizer.CurrentToken
	p.tokenizer.Next()
	right := p.expression(accept)
	return &ExpressionBinaryExpression{
		Operator: operatorBinaryMap[t.Type],
		Left:     left,
		Right:    right,
	}
}

func (p *Parser) sequenceExpression(left Expression) *ExpressionSequenceExpression {
	list := []Expression{left}
	for {
		t := p.tokenizer.CurrentToken
		if t.Type != TComma {
			break
		}
		p.tokenizer.Next()
		accept := p.acceptContextLowest()
		expr := p.expression(accept)
		list = append(list, expr)
	}
	return &ExpressionSequenceExpression{
		Expressions: list,
	}
}

func (p *Parser) conditionalExpression(left Expression, accept *acceptContext) *ExpressionConditionalExpression {
	p.tokenizer.MustMatch(TQuestion)
	consequent := p.expression(accept)
	p.tokenizer.MustMatch(TColon)
	alternate := p.expression(accept)
	p.automaticSemicolonInsertion()
	return &ExpressionConditionalExpression{
		Test:       left,
		Consequent: consequent,
		Alternate:  alternate,
	}
}

func (p *Parser) logicalExpression(left Expression, accept *acceptContext) *ExpressionLogicalExpression {
	t := p.tokenizer.CurrentToken
	tokenTypes := []TokenType{
		TAmpersandAmpersand,
		TPipePipe,
		TQuestionQuestion,
	}
	if lo.Contains(tokenTypes, t.Type) {
		p.tokenizer.Next()
		right := p.expression(accept)
		return &ExpressionLogicalExpression{
			Operator: operatorLogicalMap[t.Type],
			Left:     left,
			Right:    right,
		}
	}
	panic("logicalExpression: unexpected token")
}

func (p *Parser) equalityExpression(left Expression, accept *acceptContext) *EqualityExpression {
	t := p.tokenizer.CurrentToken
	tokenTypes := []TokenType{
		TEqualsEquals,
		TNotEquals,
		TStrictEquals,
		TStrictNotEquals,
	}
	if lo.Contains(tokenTypes, t.Type) {
		p.tokenizer.Next()
		right := p.expression(accept)
		return &EqualityExpression{
			Operator: operatorEqualityMap[t.Type],
			Left:     left,
			Right:    right,
		}
	}
	panic("equalityExpression: unexpected token")
}

func (p *Parser) relationalExpression(left Expression, accept *acceptContext) *RelationalExpression {
	t := p.tokenizer.CurrentToken
	tokenTypes := []TokenType{
		TLessThan,
		TLessThanEquals,
		TGreaterThan,
		TGreaterThanEquals,
		TInstanceof,
		TIn,
	}
	if lo.Contains(tokenTypes, t.Type) {
		p.tokenizer.Next()
		right := p.expression(accept)
		return &RelationalExpression{
			Operator: operatorRelationMap[t.Type],
			Left:     left,
			Right:    right,
		}
	}
	panic("relationalExpression: unexpected token")
}

func (p *Parser) memberExpression(left Expression) *MemberExpression {
	token := p.tokenizer.CurrentToken
	var property ASTProperty
	if token.Type == TLeftBracket {
		p.tokenizer.Next()
		propertyExpression := p.expression(p.acceptContextLowest())
		p.tokenizer.MustMatch(TRightBracket)
		property = &ASTPropertyExpression{
			Expression: propertyExpression,
		}
	} else if token.Type == TDot {
		p.tokenizer.Next()
		identifier := p.tokenizer.CurrentToken
		if identifier.Type != TIdentifier {
			// keyword after dot is treated as identifier
			if _, ok := keywordsMap[identifier.Value]; !ok {
				panic("memberExpression: expected identifier")
			}
		}
		p.tokenizer.Next()
		property = &ASTPropertyIdentifier{
			Identifier: IdentifierName(identifier.Value),
		}
	} else {
		panic("memberExpression: unexpected token")
	}

	return &MemberExpression{
		Member:   left,
		Property: property,
	}
}

func (p *Parser) superProperty() (sp SuperProperty, ok bool) {
	if !p.inClassBody || p.inMethodDefinition {
		return
	}
	p.tokenizer.Match(TSuper)
	t := p.tokenizer.CurrentToken
	switch t.Type {
	case TDot:
		p.tokenizer.Next()
		identifier := p.tokenizer.CurrentToken
		return &SuperPropertyIdentifier{
			IdentifierName: IdentifierName(identifier.Value),
		}, true
	case TLeftBracket:
		expr := p.expression(p.acceptContextLowest())
		p.tokenizer.Match(TRightBracket)
		return &SuperPropertyExpression{
			Expression: expr,
		}, true
	default:
		return
	}
}

func (p *Parser) callExpression(left Expression) *CallExpression {
	args := p.arguments()
	return &CallExpression{
		Callee:    left,
		Arguments: args,
	}
}

func (p *Parser) arguments() Arguments {
	p.tokenizer.MustMatch(TLeftParen)
	var list Arguments
	for {
		t := p.tokenizer.CurrentToken
		if t.Type == TRightParen {
			break
		}
		expr := p.expression(p.acceptContextHigherThan(TComma))
		p.tokenizer.Match(TComma)
		list = append(list, expr)
	}
	p.tokenizer.MustMatch(TRightParen)
	return list
}

func (p *Parser) arrowFunction() *ArrowFunction {
	startOffset := p.tokenizer.Index
	p.tokenizer.Match(TLeftParen)
	params := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TArrow)
	p.noLineTerminatorHere()
	var body *FunctionBody
	if p.tokenizer.Match(TLeftBrace) {
		body = p.functionBody(FunctionTypeNormal)
		p.tokenizer.MustMatch(TRightBrace)
	} else {
		expression := p.expression(p.acceptContext(TComma))
		body = &FunctionBody{
			StatementList: StatementList{
				&StatementListItemStatement{
					Statement: &StatementReturn{
						Expression: expression,
					},
				},
			},
		}
	}
	return &ArrowFunction{
		FormalParameters: params,
		Body:             body,
		SourceText:       p.SourceText[startOffset:p.tokenizer.Index],
	}
}

func (p *Parser) parenthesizedExpression() *PrimaryExpressionParenthesizedExpression {
	p.tokenizer.MustMatch(TLeftParen)
	expr := p.expression(p.acceptContextLowest())
	p.tokenizer.MustMatch(TRightParen)
	return &PrimaryExpressionParenthesizedExpression{
		Expression: expr,
	}
}

func (p *Parser) identifierReference() *PrimaryExpressionIdentifierReference {
	t := p.tokenizer.CurrentToken
	types := []TokenType{TIdentifier, TAwait, TYield}
	if !lo.Contains(types, t.Type) {
		panic("identifierReference: expected identifierOrKeyword")
	}
	name := t.Value
	p.tokenizer.Next()
	return &PrimaryExpressionIdentifierReference{
		Identifier: IdentifierName(name),
	}
}

func (p *Parser) bindingIdentifier() IdentifierName {
	t := p.tokenizer.CurrentToken
	types := []TokenType{TIdentifier, TAwait, TYield}
	if !lo.Contains(types, t.Type) {
		panic("identifierReference: expected identifierOrKeyword")
	}
	name := t.Value
	p.tokenizer.Next()
	return IdentifierName(name)
}

func (p *Parser) primaryExpression() PrimaryExpression {
	t := p.tokenizer.CurrentToken
	switch t.Type {
	case TThis:
		p.tokenizer.Next()
		return &PrimaryExpressionThis{}
	case TLeftBracket:
		return p.arrayLiteral()
	case TLeftBrace:
		return p.objectLiteral()
	case TIdentifier:
		return p.identifierReference()
	case TLeftParen:
		if e, ok := parserRecoverOk(p, p.arrowFunction); ok {
			return e
		}
		return p.parenthesizedExpression()
	case TFunction:
		if p.tokenizer.NextToken.Type == TStar {
			return p.generatorExpression()
		}
		return p.functionExpression()
	case TAsync:
		p.tokenizer.MustMatch(TAsync)
		if p.tokenizer.NextToken.Type == TStar {
			return p.asyncGeneratorExpression()
		}
		if p.tokenizer.CurrentToken.Type == TLeftParen {
			return p.asyncArrowFunction()
		}
		return p.asyncFunctionExpression()
	case TRegularExpression:
		return p.regularExpressionLiteral()
	case TClass:
		return p.classExpression()
	case TTemplateHead, TNoSubstitutionTemplate:
		return p.templateLiteral()
	default:
		literal := p.literal()
		return &PrimaryExpressionLiteral{
			Literal: literal,
		}
	}
}

func (p *Parser) templateLiteral() *PrimaryExpressionTemplateLiteral {
	if p.tokenizer.CurrentToken.Type == TNoSubstitutionTemplate {
		text := p.tokenizer.CurrentToken.Value
		p.tokenizer.Next()
		return &PrimaryExpressionTemplateLiteral{
			TemplateLiteral: &TemplateLiteral{
				Spans: []*TemplateSpan{
					{
						Text:       text,
						Expression: nil,
					},
				},
			},
		}
	}
	startIndex := p.tokenizer.CurrentStartIndex()
	templateHead := p.tokenizer.CurrentToken
	p.tokenizer.MustMatch(TTemplateHead)
	var spans []*TemplateSpan
	var expr Expression
	for {
		t := p.tokenizer.CurrentToken
		if t.Type == TTemplateTail {
			spans = append(spans, &TemplateSpan{
				Text:       t.Value,
				Expression: expr,
			})
			p.tokenizer.Next()
			break
		}
		if t.Type == TTemplateMiddle {
			spans = append(spans, &TemplateSpan{
				Text:       t.Value,
				Expression: expr,
			})
			p.tokenizer.Next()
			continue
		}
		expr = p.expression(p.acceptContextLowest())
	}
	sourceText := p.SourceText[startIndex:p.tokenizer.CurrentStartIndex()]
	return &PrimaryExpressionTemplateLiteral{
		TemplateLiteral: &TemplateLiteral{
			TemplateHead: &TemplateSpan{
				Text: templateHead.Value,
			},
			Spans: spans,
		},
		SourceText: sourceText,
	}
}

func (p *Parser) classExpression() *PrimaryExpressionClassExpression {
	startIndex := p.tokenizer.CurrentStartIndex()
	p.tokenizer.MustMatch(TClass)
	var identifier IdentifierName
	if p.tokenizer.Match(TIdentifier) {
		identifier = p.bindingIdentifier()
	}
	classTail := p.classTail()
	sourceText := p.SourceText[startIndex:p.tokenizer.CurrentStartIndex()]
	return &PrimaryExpressionClassExpression{
		IdentifierName: identifier,
		ClassTail:      classTail,
		SourceText:     sourceText,
	}
}

func (p *Parser) regularExpressionLiteral() *PrimaryExpressionRegularExpressionLiteral {
	t := p.tokenizer.CurrentToken
	p.tokenizer.MustMatch(TRegularExpression)
	var flags string
	if p.tokenizer.CurrentToken.Type == TIdentifier {
		flags = p.tokenizer.CurrentToken.Value
		p.tokenizer.Next()
	}
	return &PrimaryExpressionRegularExpressionLiteral{
		Pattern: t.Value,
		Flags:   flags,
	}
}

func (p *Parser) objectLiteral() *PrimaryExpressionObjectLiteral {
	p.tokenizer.MustMatch(TLeftBrace)
	list := p.propertyDefinitionList()
	p.tokenizer.MustMatch(TRightBrace)
	return &PrimaryExpressionObjectLiteral{
		PropertyList: list,
	}
}

func (p *Parser) propertyDefinitionList() *PropertyDefinitionList {
	var list []PropertyDefinition
	for {
		t := p.tokenizer.CurrentToken
		if t.Type == TRightBrace {
			break
		}
		if t.Type == TComma {
			p.tokenizer.Next()
			continue
		}
		prop := p.propertyDefinition()
		list = append(list, prop)
		if p.tokenizer.CurrentToken.Type == TComma {
			p.tokenizer.Next()
		}
	}
	return &PropertyDefinitionList{
		Items: list,
	}
}

func (p *Parser) propertyName() (PropertyName, bool) {
	var propertyName PropertyName
	t := p.tokenizer.CurrentToken
	switch t.Type {
	case TIdentifier:
		identifierRef := p.identifierReference()
		propertyName = &PropertyNameLiteralIdentifier{
			Identifier: identifierRef.Identifier,
		}
	case TString:
		stringLiteral := p.stringLiteral()
		propertyName = &PropertyNameLiteralString{
			StringLiteral: stringLiteral,
		}
	case TNumber, TBigInt:
		numberLiteral := p.numericLiteral()
		propertyName = &PropertyNameLiteralNumeric{
			NumericLiteral: numberLiteral,
		}
	case TLeftBracket:
		p.tokenizer.Next()
		computedPropertyName := p.expression(p.acceptContextLowest())
		propertyName = &ComputedPropertyName{
			Expression: computedPropertyName,
		}
		p.tokenizer.MustMatch(TRightBracket)
	default:
		return nil, false
	}
	return propertyName, true
}

func (p *Parser) methodDefinition(methodType MethodDefinitionType) *MethodDefinition {
	p.tokenizer.store()
	inMethodDefinition := p.inMethodDefinition
	inClassConstructor := p.inClassConstructor
	p.inMethodDefinition = true
	p.inClassConstructor = true
	defer func() {
		p.inMethodDefinition = inMethodDefinition
		p.inClassConstructor = inClassConstructor
	}()
	if methodType == MethodDefinitionTypeNil {
		if p.tokenizer.Match(TStar) {
			return p.methodDefinition(MethodDefinitionTypeGenerator)
		}
	}
	propertyName, ok := p.propertyName()
	if methodType == MethodDefinitionTypeNil && ok {
		literal, ok := propertyName.(LiteralPropertyName)
		if ok {
			isGet := literal.LiteralString() == "get"
			isSet := literal.LiteralString() == "set"
			if literal.LiteralString() == "async" {
				if p.tokenizer.CurrentToken.Type == TStar {
					return p.methodDefinition(MethodDefinitionTypeAsyncGenerator)
				}
				return p.methodDefinition(MethodDefinitionTypeAsync)
			}
			if isGet {
				return p.methodDefinition(MethodDefinitionTypeGet)
			}
			if isSet {
				return p.methodDefinition(MethodDefinitionTypeSet)
			}
		}
	}
	p.tokenizer.Match(TLeftParen)
	start := p.tokenizer.CurrentStartIndex()
	formalParameters := p.formalParameters()
	p.tokenizer.MustMatch(TRightParen)
	p.tokenizer.MustMatch(TLeftBrace)
	body := p.functionBody(FunctionTypeNormal)
	p.tokenizer.MustMatch(TRightBrace)
	sourceText := p.SourceText[start:p.tokenizer.CurrentStartIndex()]
	m := MethodDefinitionTypeMethod
	if methodType != MethodDefinitionTypeNil {
		m = methodType
	}
	var funExpression *PrimaryExpressionFunctionExpression
	if m == MethodDefinitionTypeMethod || m == MethodDefinitionTypeGet || m == MethodDefinitionTypeSet {
		funExpression = &PrimaryExpressionFunctionExpression{
			Identifier:       "",
			FormalParameters: formalParameters,
			SourceText:       sourceText,
			Body:             body,
		}
	}
	var genExpression *PrimaryExpressionGeneratorExpression
	if m == MethodDefinitionTypeGenerator {
		genExpression = &PrimaryExpressionGeneratorExpression{
			IdentifierName:   "",
			FormalParameters: formalParameters,
			SourceText:       sourceText,
			Body:             body,
		}
	}
	var asyncExpression *PrimaryExpressionAsyncFunctionExpression
	if m == MethodDefinitionTypeAsync {
		asyncExpression = &PrimaryExpressionAsyncFunctionExpression{
			Identifier:       "",
			FormalParameters: formalParameters,
			SourceText:       sourceText,
			Body:             body,
		}
	}
	var asyncGenerator *PrimaryExpressionAsyncGeneratorExpression
	if m == MethodDefinitionTypeAsyncGenerator {
		asyncGenerator = &PrimaryExpressionAsyncGeneratorExpression{
			IdentifierName:   "",
			FormalParameters: formalParameters,
			SourceText:       sourceText,
			Body:             body,
		}
	}
	return &MethodDefinition{
		Type:                     m,
		PropertyName:             propertyName,
		FunctionExpression:       funExpression,
		GeneratorExpression:      genExpression,
		AsyncFunctionExpression:  asyncExpression,
		AsyncGeneratorExpression: asyncGenerator,
	}
}

func (p *Parser) propertyDefinition() PropertyDefinition {
	t := p.tokenizer.CurrentToken
	var propertyName PropertyName
	p.tokenizer.store()

	accept := p.acceptContextHigherThan(TComma)
	switch t.Type {
	case TDotDotDot:
		p.tokenizer.Next()
		expr := p.expression(accept)
		return &PropertyDefinitionSpread{
			Spread: expr,
		}
	default:
		if method, ok := parserRecoverOk(p, func() *MethodDefinition {
			return p.methodDefinition(MethodDefinitionTypeNil)
		}); ok {
			return &PropertyDefinitionMethodDefinition{
				method,
			}
		}
	}
	propertyName, ok := p.propertyName()
	Assert(ok)

	if p.tokenizer.Match(TColon) {
		value := p.expression(accept)
		return &PropertyDefinitionNameAndExpression{
			Name:       propertyName,
			Expression: value,
		}
	} else {
		identifier, ok := propertyName.(*PropertyNameLiteralIdentifier)
		Assert(ok)
		return &PropertyDefinitionIdentifierReference{
			IdentifierReference: &PrimaryExpressionIdentifierReference{
				Identifier: identifier.Identifier,
			},
		}
	}
}

func (p *Parser) arrayLiteral() *ArrayLiteral {
	p.tokenizer.MustMatch(TLeftBracket)
	var list []ArrayElement
	accept := p.acceptContext(TComma)
	for {
		t := p.tokenizer.CurrentToken
		if t.Type == TRightBracket {
			p.tokenizer.Next()
			break
		}
		if p.tokenizer.Match(TComma) {
			list = append(list, &ArrayElementElision{})
		} else if p.tokenizer.Match(TDotDotDot) {
			expr := p.expression(accept)
			list = append(list, &ArrayElementSpread{
				Spread: expr,
			})
			p.tokenizer.Match(TComma)
		} else {
			expr := p.expression(accept)
			list = append(list, &ArrayElementExpression{
				Expression: expr,
			})
			p.tokenizer.Match(TComma)
		}
	}
	return &ArrayLiteral{
		ElementList: list,
	}
}

func (p *Parser) literal() Literal {
	t := p.tokenizer.CurrentToken
	defer p.tokenizer.Next()
	switch t.Type {
	case TTrue, TFalse:
		return &LiteralBoolean{
			Bool: t.Type == TTrue,
		}
	case TNull:
		return &LiteralNull{}
	case TUndefined:
		return &LiteralUndefined{}
	case TNumber, TBigInt:
		return p.numericLiteral()
	case TString:
		return p.stringLiteral()
	case TComment:
		return &LiteralUndefined{}
	default:
		panic("literal: unhandled token")
	}
}

func (p *Parser) numericLiteral() *LiteralNumeric {
	t := p.tokenizer.CurrentToken
	var numericType NumericType
	if t.Type == TBigInt {
		numericType = NumericTypeBigInt
	} else {
		numericType = NumericTypeNumber
	}
	return &LiteralNumeric{
		Value: t.Value,
		Type:  numericType,
	}
}

func (p *Parser) stringLiteral() *LiteralString {
	t := p.tokenizer.CurrentToken
	if t.Type != TString {
		panic("stringLiteral: expected string")
	}
	return &LiteralString{
		Value: t.Value,
	}
}

func (p *Parser) variableStatement() *StatementVariable {
	p.tokenizer.MustMatch(TVar)
	list := p.variableDeclarationList()
	return &StatementVariable{
		DeclarationList: list,
	}
}

func (p *Parser) variableDeclarationList() *VariableDeclarationList {
	var list []*VariableDeclaration
	for {
		if declaration, ok := parserRecoverOk(p, p.variableDeclaration); ok {
			list = append(list, declaration)
		} else {
			break
		}
		if p.tokenizer.CurrentToken.Type == TComma {
			p.tokenizer.Next()
			continue
		}
		if p.tokenizer.CurrentToken.Type == TSemicolon {
			p.tokenizer.Next()
			break
		}
		if p.tokenizer.CurrentToken.Type == TEOF {
			break
		}
	}
	return &VariableDeclarationList{Items: list}
}

func (p *Parser) variableDeclaration() *VariableDeclaration {
	identifier := p.bindingIdentifier()
	var init Expression
	if p.tokenizer.Match(TEquals) {
		init = p.expression(p.acceptContextHigherThan(TComma))
	}
	return &VariableDeclaration{
		BindingIdentifier: identifier,
		Initializer:       init,
	}
}

func (p *Parser) blockStatement() *BlockStatementBlock {
	return &BlockStatementBlock{
		Block: p.block(),
	}
}

func (p *Parser) block() *Block {
	p.tokenizer.MustMatch(TLeftBrace)
	list := p.statementList()
	p.tokenizer.MustMatch(TRightBrace)

	return &Block{
		StatementList: list,
	}
}

func parserRecoverOk[T any](p *Parser, f func() T) (r T, ok bool) {
	p.tokenizer.store()
	defer func() {
		if r := recover(); r != nil {
			// fmt.Println("parser recovered from: ", pkg.GetFunctionName(f), r)
			p.tokenizer.restore()
			ok = false
		} else {
			p.tokenizer.popCachedState()
		}
	}()
	return f(), true
}
