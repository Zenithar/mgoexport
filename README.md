# mgoexport
Tool to export MongoDB Collection as CSV files (Golang)

## Usage

    mgoexport -db "test" -c "test" -o "test.csv"

Export ALL columns of ALL entries of the collection in the given CSV file

    mgoexport -H 127.0.0.1:27017 -db "remote_test" -c "test" -o "remote.csv"

Same effect as the previous command but connect to a remote server

    mgoexport -db "test" -c "test" -o "test.csv" -f "_id,range.to"

Only export fields named "_id", and subfield "range.to"

WARNING: No space in the fieldset.

## Inspirations & Credits

[json2csv](https://github.com/jehiah/json2csv/)

[Kelley Robinson](http://grokbase.com/user/Kelley-Robinson/ckVVvA1fszm6sqXXLF2wqR) for the flattening algorithm
