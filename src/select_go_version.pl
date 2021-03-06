#!/usr/bin/env perl
use strict;
use warnings;

use FindBin 1.51 qw( $RealBin );
use lib $RealBin;

require "lib.pl";

my $usage = "usage: echo <versions...> | select_go_version.pl pattern";

my $pat_arg = shift;
exit_err($usage) unless $pat_arg;
my $pat = parse_go_version($pat_arg);
exit_err("invalid pattern: $pat_arg") unless $pat;

my $max;

foreach my $v (<STDIN>) {
    chomp($v);
    next unless my $pv = parse_go_version($v);
    next unless go_version_pattern_match( $pat, $pv );
    if ( go_version_greater( $pv, $max ) ) {
        $max = $pv;
    }
}

exit 1 unless $max;
print go_version_string($max) . "\n";
