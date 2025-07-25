### RoadMap.. 

### As this tool would be implemented in stages defined below. 

| Phase          | Key Tasks                          | Tools/Techniques              | Expected Output                          |
|----------------|-----------------------------------|-------------------------------|------------------------------------------|
| **Setup**      | Define API, handle `null`/booleans | Python functions, exceptions  | `parse("true") → True`                   |
| **Tokenizer**  | Split JSON into tokens (`{`, `"a"`)| Regex, string manipulation    | `tokenize('{"x":1}') → ['{','"x"',...]` |
| **Parser**     | Convert tokens to Python objects  | Recursive descent parsing     | `parse('{"x":1}') → {"x": 1}`           |
| **Error Handling** | Catch malformed JSON          | Custom exceptions (`JSONDecodeError`) | `parse("{") → Error`       |
| **Advanced**   | Unicode escapes, scientific numbers | String processing, `float()`  | `parse('"\\u263A"') → "☺"`              |
| **Testing**    | Validate correctness & edge cases  | `unittest`, Hypothesis fuzzing | 100% test coverage                      |