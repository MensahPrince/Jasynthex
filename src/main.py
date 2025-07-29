import re

#A function to convert literals by taking in the stringsets, true, false or null and returning the appropriate 
# literals: True, False, None.

def convert_literal(token):
    if token == "true":
        return True
    if token == "false":
        return False
    if token == "null":
        return None
    
    #Checks if there is a . in a token and makes it converts it into a float. Else, it is an int 
    try:
        if '.' in token:
            return float(token)
        else:
            return int(token)
    except ValueError:
        return token

#Function to tokenize the json string. 

def tokenize(json_str):

    #Regular expressions to identify basic json objects, these are Whitespaces, Strings, Numbers, Booleans, and the structure of the json
    token_spec = r'''
        (?P<WHITESPACE>\s+)                             |
        (?P<STRING>"(\\.|[^"\\])*")                     |
        (?P<NUMBER>-?\d+(\.\d+)?([eE][+-]?\d+)?)         |
        (?P<BOOLEAN>true|false|null)                    |
        (?P<STRUCTURE>[{}\[\]:,])                       
    '''
    token_regex = re.compile(token_spec, re.VERBOSE)

    #array to contain the tokinized tokens from the json string. 
    tokens = []
    
    #check for the presence of the defined regex 
    for match in token_regex.finditer(json_str):
        kind = match.lastgroup
        value = match.group()
        if kind == "WHITESPACE":
            continue
        tokens.append(value)
    return tokens


# Function to parse tokens and return python object, 
def parse(tokens):
    def parse_value():
        nonlocal index
        token = tokens[index]

        if token == '{':
            index += 1
            obj = {}
            while tokens[index] != '}':
                key = parse_value()
                index += 1  # skip colon
                value = parse_value()
                obj[key] = value
                if tokens[index] == ',':
                    index += 1
            index += 1  # skip closing }
            return obj

        elif token == '[':
            index += 1
            arr = []
            while tokens[index] != ']':
                arr.append(parse_value())
                if tokens[index] == ',':
                    index += 1
            index += 1  # skip closing ]
            return arr

        elif token.startswith('"') and token.endswith('"'):
            index += 1
            return token[1:-1]

        else:
            index += 1
            return convert_literal(token)

    index = 0
    return parse_value()


# example json input string
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
parsed = parse(tokens)
print(parsed)
