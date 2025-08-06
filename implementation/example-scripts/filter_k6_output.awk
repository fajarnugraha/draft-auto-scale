#!/usr/bin/awk -f

# This script filters the verbose output from the k6 load testing tool.
# It identifies and captures the final summary block while ignoring the
# repetitive progress updates. It will also print any lines containing
# "ERRO" or "WARN" to ensure important messages are not missed.

# Skip the initial "scenarios" block and other startup messages
/scenarios: \(/ {
  in_header = 1
}

# The summary block starts with a bold escape code.
# When we see this, start capturing lines.
/^\x1b\[1m/ {
  in_summary = 1
}

# If we are in the summary block, append the current line.
in_summary {
  summary = summary $0 "\n"
}

# Always print lines with errors or warnings.
/ERRO|WARN/ {
  print
}

# END block: Executed after all lines are processed.
# Print the captured summary.
END {
  if (summary) {
    print summary
  }
}
