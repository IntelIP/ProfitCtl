import assert from "node:assert/strict";
import { calculateExample } from "../src/scripts/demo.ts";

const feature = calculateExample({ users: 1000, requests: 60, price: 20 });
assert.equal(feature.revenue, 20000);
assert.equal(feature.baseline.cost, 2500);
assert.equal(feature.candidate.cost, 6700);
assert.equal(feature.candidate.margin, 66.5);
assert.equal(feature.addedCost, 4200);
assert.equal(feature.minimumPrice, 16.75);
assert.equal(feature.passes, true);

const pricing = calculateExample({ users: 1000, requests: 60, price: 12 });
assert.ok(Math.abs(pricing.candidate.margin - 44.1666666667) < 1e-8);
assert.equal(pricing.passes, false);
const growth = calculateExample({ users: 10000, requests: 220, price: 20 });
assert.equal(growth.candidate.cost, 191400);
assert.equal(growth.candidate.margin, 4.3);
assert.equal(growth.passes, false);
assert.equal(calculateExample({ users: 1000, requests: 0, price: 20 }).addedCost, 0);
for (const bad of [0, -1, NaN, Infinity]) {
  assert.throws(() => calculateExample({ users: bad, requests: 60, price: 20 }), RangeError);
  assert.throws(() => calculateExample({ users: 1000, requests: 60, price: bad }), RangeError);
}
assert.throws(() => calculateExample({ users: 1000, requests: -1, price: 20 }), RangeError);
console.log("Interactive example: costs, margin decisions, and invalid inputs verified.");
