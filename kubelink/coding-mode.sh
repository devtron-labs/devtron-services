#!/bin/bash

# Check if the correct argument is passed
if [[ "$1" != "prod" && "$1" != "dev" ]]; then
    echo "Error: Argument not supported. Use 'prod' or 'dev'."
    exit 1
fi

# Input and output files
input_file="go.mod"
temp_file=$(mktemp)

# Flags for identifying blocks
in_block=false
below_dev=false

# Read through the file line by line
while IFS= read -r line; do

    # Detect the start of the block (//prod)
    if [[ "$line" == *"//prod"* ]]; then
        in_block=true
        echo "$line" >> "$temp_file"
        continue
    fi

    # Detect the end of the block (//dev)
    if [[ "$line" == *"//dev"* ]]; then
        in_block=false
        below_dev=true
        echo "$line" >> "$temp_file"
        continue
    fi

    # Behavior when "prod" is passed
    if [[ "$1" == "prod" ]]; then
        # If within the block (between //prod and //dev), uncomment lines
        if [ "$in_block" = true ]; then
            uncommented_line=$(echo "$line" | sed 's|^// ||')
            echo "$uncommented_line" >> "$temp_file"
        # Comment all lines below //dev if not already commented
        elif [ "$below_dev" = true ]; then
            if [[ "$line" != "// "* ]]; then
                echo "// $line" >> "$temp_file"
            else
                echo "$line" >> "$temp_file"
            fi
        else
            echo "$line" >> "$temp_file"
        fi
    fi

    # Behavior when "dev" is passed
    if [[ "$1" == "dev" ]]; then
        # If within the block (between //prod and //dev), comment lines if not already commented
        if [ "$in_block" = true ]; then
            if [[ "$line" != "// "* ]]; then
                echo "// $line" >> "$temp_file"
            else
                echo "$line" >> "$temp_file"
            fi
        # Uncomment all lines below //dev
        elif [ "$below_dev" = true ]; then
            uncommented_line=$(echo "$line" | sed 's|^// ||')
            echo "$uncommented_line" >> "$temp_file"
        else
            echo "$line" >> "$temp_file"
        fi
    fi

done < "$input_file"

# Move the temporary file to overwrite the original go.mod file
mv "$temp_file" "$input_file"

echo "Successfully updated go.mod based on the '$1' argument. Now running go mods and make"

go mod tidy

go mod vendor

make

