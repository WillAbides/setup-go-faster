#!/usr/bin/env perl
use strict;
use warnings;

use FindBin 1.51 qw( $RealBin );
use lib $RealBin;

require "lib.pl";

my $usage = "usage: echo <versions...> | select_go_version.pl pattern";

my $pat_arg = shift;
unless ($pat_arg) {
    print "$usage\n";
    exit 1;
}

my $pat = parse_go_version($pat_arg);
unless ($pat) {
    print "invalid pattern: $pat_arg\n";
    exit 1;
}

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

print $$max{"original"} . "\n";
