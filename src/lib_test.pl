#!/usr/bin/env perl
use strict;
use warnings;
use Test::More;
use FindBin 1.51 qw( $RealBin );
use lib $RealBin;

require "lib.pl";

sub parse_and_greater {
    my $a = parse_go_version(shift);
    my $b = parse_go_version(shift);
    go_version_greater( $a, $b );
}

ok( parse_and_greater( "1.0.1",     "1.0.0" ) == 1 );
ok( parse_and_greater( "1.0.0",     "1.0.0" ) == 0 );
ok( parse_and_greater( "1.0.0",     "1.0.1" ) == 0 );
ok( parse_and_greater( "1.12",      "1.11.1" ) == 1 );
ok( parse_and_greater( "1.12.8",    "1.13.7" ) == 0 );
ok( parse_and_greater( "1.12beta1", "1.12beta2" ) == 0 );
ok( parse_and_greater( "1.12beta2", "1.12beta1" ) == 1 );
ok( parse_and_greater( "1.12",      "1.12beta1" ) == 1 );
ok( parse_and_greater( "1.12beta1", "1.12" ) == 0 );
ok( parse_and_greater( "1.12.0",    "1.12" ) == 0 );
ok( parse_and_greater( "1.12",      "1.12.0" ) == 0 );

sub parse_and_match {
    my $a = parse_go_version(shift);
    my $b = parse_go_version(shift);
    go_version_pattern_match( $a, $b );
}

ok( parse_and_match( "x",     "1.2.3" ) == 1 );
ok( parse_and_match( "1.x",   "1.2.3" ) == 1 );
ok( parse_and_match( "1.2.x", "1.2.3" ) == 1 );
ok( parse_and_match( "1.2.x", "1.2.3beta1" ) == 0 );

sub test_go_version_string {
    my $in = shift;
    my $p  = parse_go_version($in);
    ok( go_version_string($p) eq $in, "test_go_version_string $in" );
}

test_go_version_string "1.2.3";
test_go_version_string "1";
test_go_version_string "0";
test_go_version_string "1.0.1";
test_go_version_string "1.1";
test_go_version_string "1.1beta1";
test_go_version_string "1beta1";

done_testing();

