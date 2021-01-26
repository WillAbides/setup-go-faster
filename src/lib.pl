#!/usr/bin/perl
use strict;
use warnings;

my $exp = qr/^(x|\d+)(\.(x|\d+)(\.(x|\d+))?)?(\w+)?$/;

sub parse_go_version {
  my $ver = shift;
  return unless $ver =~ m/$exp/;
  my $major = $1;
  my $minor = $3;
  my $patch = $5;
  my $pre_release = $6;
  $major = 0 unless $major;
  $minor = 0 unless $minor;
  $patch = 0 unless $patch;
  $pre_release = "" unless $pre_release;
  my $result = { major => $major, minor => $minor, patch => $patch, pre_release => $pre_release };
  return unless is_valid_go_version_pattern($result);
  return $result;
}

sub go_version_greater {
  my $a = shift;
  my $b = shift;
  if ($a && ! $b) {
    return 1
  }
  if ($b && ! $a) {
    return 0
  }
  foreach ("major", "minor", "patch") {
    return 1 if $$a{$_} > $$b{$_};
  }
  # true if b is a pre-release and a isn't
  return 1 if $$b{"pre_release"} ne "" && $$a{"pre_release"} eq "";
  # false if a is a pre-release and b isn't
  return 0 if $$a{"pre_release"} ne "" && $$b{"pre_release"} eq "";
  # true if a's preview is asciibetical ahead of b's
  return 1 if $$a{"pre_release"} gt $$b{"pre_release"};
  return 0;
}

sub is_valid_go_version_pattern() {
  my $v = shift;
  if ($$v{"major"} eq "x") {
    return 0 unless $$v{"minor"} ne "x" && $$v{"patch"} ne "x" && $$v{"pre_release"} eq "";
  }
  if ($$v{"minor"} eq "x") {
    return 0 unless $$v{"patch"} ne "x" && $$v{"pre_release"} eq "";
  }
  if ($$v{"patch"} eq "x") {
    return 0 unless $$v{"pre_release"} eq "";
  }
  return 1;
}

sub go_version_string {
  my $v = shift;
  my $major_v = $$v{"major"};
  my $minor_v = $$v{"minor"};
  my $patch_v = $$v{"patch"};
  $patch_v = "" if $patch_v == 0;
  $minor_v = "" if $minor_v == 0 && $patch_v eq "";
  my $st = "$$v{'major'}";
  if ($minor_v ne "") {
    $st = "$st.$minor_v";
  }
  if ($patch_v ne "") {
    $st = "$st.$patch_v";
  }
  if ($$v{"pre_release"}) {
    $st = "$st$$v{'pre_release'}";
  }
  return $st;
}

sub go_version_pattern_match {
  my $pattern = shift;
  my $ver = shift;
  foreach ("major", "minor", "patch") {
    last if $$pattern{$_} eq "x";
    return 0 if $$pattern{$_} != $$ver{$_};
  }
  return 0 if $$pattern{"pre_release"} ne $$ver{"pre_release"};
  return 1;
}

sub exit_err() {
  my $msg = shift;
  print "$msg\n";
  exit 1;
}

1;
