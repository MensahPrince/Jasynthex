def tokenize(json_str):
    tokens = []
    i = 0
    length = len(json_str)

    while i < length:
        char = json_str[i]

        # Skip whitespace
        if char in ' \t\n\r':
            i += 1
            continue

        # Structural tokens
        if char in '{}[]:,':
            tokens.append(char)
            i += 1
            continue

        # Strings
        # Handle strings with escaped quotes
        if char == '"':
            i += 1
            start = i
            while i < length:
                if json_str[i] == '"' and json_str[i - 1] != '\\':
                    break
                i += 1
            tokens.append('"' + json_str[start:i] + '"')
            i += 1
            continue

        # Numbers
        if char.isdigit() or char == '-':
            start = i
            while i < length and (json_str[i].isdigit() or json_str[i] in '.eE+-'):
                i += 1
            tokens.append(json_str[start:i])
            continue

        # Literals: true, false, null
        if json_str[i:i+4] == "true":
            tokens.append("true")
            i += 4
            continue
        if json_str[i:i+5] == "false":
            tokens.append("false")
            i += 5
            continue
        if json_str[i:i+4] == "null":
            tokens.append("null")
            i += 4
            continue

        # Catch any unexpected input
        raise ValueError(f"Unexpected character at position {i}: {char}")

    return tokens

# Test the tokenizer
json_input = '''
{
    "name": "Jasynthex",
    "version": 1.0,
    "active": true,
    "features": ["tokenizer", "parser", null],
    "metadata": {
        "created_by": "MensahPrince",
        "license": null
    }
}
'''

tokens = tokenize(json_input)
print(tokens)
