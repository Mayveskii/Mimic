# Structured Outputs

Return machine-readable JSON from chat completions.

Use `response_format: { "type": "json_object" }` for most structured outputs in GonkaGate chat completions. Switch to `json_schema` only when your client depends on a stricter response contract and the selected model/provider pair supports it.

## Quick start

Set `response_format` to `json_object`, tell the model to return JSON only, then parse `choices[0].message.content` in your app.
 Quick start
```
export GONKAGATE_API_KEY="gp-your-api-key"

curl https://api.gonkagate.com/v1/chat/completions \
  -H "Authorization: Bearer $GONKAGATE_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "qwen/qwen3-235b-a22b-instruct-2507-fp8",
    "messages": [
      { "role": "system", "content": "Return JSON only. Output must be valid JSON." },
      { "role": "user", "content": "Extract: name (string) and age (integer). Text: Alice is 30." }
    ],
    "response_format": { "type": "json_object" },
    "temperature": 0
  }'
```

Expected client value after parsing:
 Expected client value after parsing
```
{ "name": "Alice", "age": 30 }
```

In the API response, `choices[0].message.content` is still a string. Parse it in your client before validation.

## Choose `json_object` or `json_schema`

- Parseable JSON and client-side validation: use `json_object`.
- Required fields, enums, nested objects, or `additionalProperties: false`: use `json_schema`.
- Extra hardening for non-stream structured responses: add `response-healing` after `response_format` is already correct.
- The model should call your backend or another external system: use [Tool Calling](./tool-calling) instead of structured outputs.

## Switch to `json_schema` only when the contract matters

Use `json_schema` for required fields, enums, nested objects, and `additionalProperties: false`. Stay with `json_object` when parseable JSON plus client-side validation is enough.

Replace the `response_format` field in the request above with:
**Request Example**
```
{
  "type": "json_schema",
  "json_schema": {
    "name": "person",
    "strict": true,
    "schema": {
      "type": "object",
      "properties": {
        "name": { "type": "string", "description": "Person name" },
        "age": { "type": "integer", "description": "Age in years" }
      },
      "required": ["name", "age"],
      "additionalProperties": false
    }
  }
}
```

If strict schema mode is not supported by the selected model/provider pair, fall back to `json_object` and validate after parsing.

## Validate model output in your client

Treat model output as untrusted input, even when it parses successfully.

### Python (Pydantic)
 Python (Pydantic)
```
from pydantic import BaseModel

class Person(BaseModel):
    name: str
    age: int

person = Person.model_validate_json(response.choices[0].message.content)
print(person)
```

### TypeScript (Zod)
 TypeScript (Zod)
```
import { z } from "zod";

const PersonSchema = z.object({
  name: z.string(),
  age: z.number().int()
});

const content = response.choices[0]?.message?.content ?? "{}";
const person = PersonSchema.parse(JSON.parse(content));
console.log(person);
```

## Troubleshooting structured outputs

| Problem | What to check |
|---|---|
| Support is unclear or inconsistent | GonkaGate forwards response_format to the selected upstream model/provider. Support depends on that exact pair, and GET /v1/models does not expose structured-output capability flags. |
| Output is not valid JSON | Tighten the system instruction and keep the word JSON in the prompt if the selected upstream requires it. |
| JSON.parse or schema validation fails | Simplify the output contract, reduce ambiguity in the prompt, or fall back from json_schema to json_object. |
| Streamed JSON looks broken mid-response | Partial chunks are not guaranteed to be valid JSON. Buffer until the final chunk, then parse. |
| JSON is truncated | Increase max_tokens or reduce the expected output size. |

## See also

- [Response Healing](./plugins/response-healing) for extra hardening on non-stream structured responses after `response_format` is already correct.
- [Tool Calling](./tool-calling) when the model should request work from your backend or another external system.
- [Chat Completion Parameters](/en/docs/api/reference/parameters) when you need the field-level `response_format` contract.