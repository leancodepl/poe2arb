#!/bin/bash
set -e

poe2arb convert io --lang en < exports/en.json > lib/l10n/app_en.arb
poe2arb convert io --lang pl --no-template < exports/pl.json > lib/l10n/app_pl.arb
flutter gen-l10n
flutter analyze