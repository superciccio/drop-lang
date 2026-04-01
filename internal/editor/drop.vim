" Vim syntax file for Drop
" Language: Drop
" Maintainer: Drop contributors

if exists("b:current_syntax")
  finish
endif

" Comments
syn match dropComment "--.*$"

" Strings (with interpolation)
syn region dropString start=/"/ end=/"/ contains=dropInterpolation
syn match dropInterpolation /\{[^}]\+\}/ contained

" Numbers
syn match dropNumber /\<\d\+\(\.\d\+\)\?\>/

" Keywords
syn keyword dropKeyword if else for in do return print and or not

" Constants
syn keyword dropConstant true false

" Web
syn keyword dropWeb serve get post put delete respond body fetch

" Data
syn keyword dropData store save load remove

" UI
syn keyword dropUI page text row each button form input submit link image

" Highlighting
hi def link dropComment   Comment
hi def link dropString    String
hi def link dropInterpolation Special
hi def link dropNumber    Number
hi def link dropKeyword   Keyword
hi def link dropConstant  Constant
hi def link dropWeb       Function
hi def link dropData      Function
hi def link dropUI        Type

let b:current_syntax = "drop"
