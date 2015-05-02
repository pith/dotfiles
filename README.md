dotfiles [![Build Status](https://travis-ci.org/pith/dotfiles.svg)](https://travis-ci.org/pith/dotfiles)
========

Managing your dotfiles around multiple machine can be a pain. That's why a lot of initiatives came up to fix this. Most of them are listed in the [Your unofficial guide to dotfiles on GitHub](https://dotfiles.github.io/). But what I miss in these projects is a portable (I mean, whithout requiring any dependency) dotfiles manager separated from the dotfiles config. I want it to be platform and shell independent. And I want to be able to pull updates from this manager whitout pulling someone else config.

That's why I started this project based on the great work of [cowboy](https://github.com/cowboy/dotfiles). The goal here is to have a single dotfiles binary written in Go, doing all the dotfiles managing stuff for you.

*State of the project*: This project is actively developed, but is not ready yet. All the contributions are welcomed.

## Install

Since it is not released yet, you can only get it from source and it requires you have a Go environment setup. The rest is pretty simple:

    go get github.com/pith/dotfiles

## Usage

*WARNING:* This might change quickly.

Create a new .dotfiles repo:

    dotfiles create

This create a `~/.dotfiles` directory with the following architecture:

    ~/.dotfiles
     |- bin
     |- conf
     |- copy
     |- init
     |- source
     |- test
     |- vendor

Add your config files in one of these directories according to what they do.

Then, when your config is ready, run the init command.

    dotfiles init

This will copy all the files in `copy` in `~/`, symlink all the files in `link` in `~/`, run the scripts in `init` and source those in `source`.

## TODOs

* One command should be enough with good default and user inputs (remove the need to specify `create` and `init`)
* Integrate git
  - Allow to clone a .dotfiles project by passing a git URL as argument to the command
  - Execute a git pull when setting up the config
  - Allow the use of branches to get config profiles

## Copyright and license
This source code is copyrighted by Pierre THIROUIN and released under the terms of the [Mozilla Public License 2.0](LICENSE).
