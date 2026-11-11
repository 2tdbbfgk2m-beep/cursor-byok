import assert from "node:assert/strict";
import test from "node:test";

import {
  buildProviderGroupKey,
  buildProviderGroupKeyFromCredentials,
  buildProviderGroups,
} from "./providerGroups.js";

test("buildProviderGroupKey normalizes baseURL case and trailing slash", () => {
  const left = buildProviderGroupKey({ baseURL: "HTTPS://API.OpenAI.com/v1/", apiKey: " sk-1 " });
  const right = buildProviderGroupKey({ baseURL: "https://api.openai.com/v1", apiKey: "sk-1" });
  assert.equal(left, right);
});

test("buildProviderGroupKey separates same baseURL with different keys", () => {
  const left = buildProviderGroupKey({ baseURL: "https://api.openai.com/v1", apiKey: "sk-1" });
  const right = buildProviderGroupKey({ baseURL: "https://api.openai.com/v1", apiKey: "sk-2" });
  assert.notEqual(left, right);
});

test("buildProviderGroupKeyFromCredentials matches adapter-derived keys", () => {
  const adapterKey = buildProviderGroupKey({ baseURL: "https://api.deepseek.com", apiKey: "sk-a" });
  const credentialKey = buildProviderGroupKeyFromCredentials("https://api.deepseek.com/", " sk-a ");
  assert.equal(adapterKey, credentialKey);
});

test("buildProviderGroups keeps first-appearance order and collects types", () => {
  const groups = buildProviderGroups([
    { baseURL: "https://a.com/v1", apiKey: "k1", type: "openai", modelID: "m1" },
    { baseURL: "https://b.com", apiKey: "k2", type: "anthropic", modelID: "m2" },
    { baseURL: "https://a.com/v1", apiKey: "k1", type: "anthropic", modelID: "m3" },
  ]);

  assert.equal(groups.length, 2);
  assert.equal(groups[0].baseURL, "https://a.com/v1");
  assert.deepEqual(groups[0].types, ["openai", "anthropic"]);
  assert.equal(groups[0].adapters.length, 2);
  assert.deepEqual(groups[1].types, ["anthropic"]);
});

test("buildProviderGroups merges URL variants of the same provider", () => {
  const groups = buildProviderGroups([
    { baseURL: "https://Aggregator.io/v1", apiKey: "k", type: "openai" },
    { baseURL: "https://aggregator.io/v1/", apiKey: "k", type: "anthropic" },
  ]);

  assert.equal(groups.length, 1);
  assert.equal(groups[0].adapters.length, 2);
});

test("buildProviderGroups returns empty list for empty input", () => {
  assert.deepEqual(buildProviderGroups([]), []);
  assert.deepEqual(buildProviderGroups(null), []);
});

test("provider names label the whole group, including models added later", () => {
  const groups = buildProviderGroups([
    { baseURL: "https://a.example/v1", apiKey: "k", modelID: "new-model" },
    { baseURL: "https://a.example/v1", apiKey: "k", providerName: " VC ", modelID: "old-model" },
    { baseURL: "https://legacy.example/v1", apiKey: "k2", modelID: "same-model" },
  ]);
  assert.equal(groups[0].name, "VC");
  assert.equal(groups[0].providerName, "VC");
  assert.equal(groups[1].name, "legacy.example");
  const renamed = buildProviderGroups(groups[0].adapters.map((a) => ({ ...a, providerName: "Renamed" })));
  assert.equal(renamed[0].name, "Renamed");
  assert.equal(renamed[0].key, groups[0].key);
});
