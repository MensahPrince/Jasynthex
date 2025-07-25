# Yello 

# Read in and assign a True, False, or None value from a JSON string

json_str = input("Enter a JSON string (true, false, null): ")

# Function to parse JSON strings and return corresponding Python values 

def parser(json_str : str):
    json_str = json_str.strip()

    if json_str == "true":
        return True
    elif json_str == "false":
        return False
    elif json_str == "null":
        return None
    else: 
        raise ValueError(f"Invalid JSON string: {json_str}")

parser(json_str)