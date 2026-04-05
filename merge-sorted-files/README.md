# my-sort

Instead of reading the whole file into memory and sorting the lines, we can do _external merge sort_. Basically, we break the original files into multiple temporary files where the size doesn't exceed a certain threshold, sort them in memory, then combine them into a result file by looking at the first line from each sorted temp files.