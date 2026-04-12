# Design-Partner Intake

This is the lightweight loop for turning a benchmark conversation into a real user relationship.

## Goal

Turn one useful benchmark run into one of three outcomes:

- an install
- a calibration request
- a design-partner pilot

For inbound evaluators from the public repo, use the GitHub `Design-partner request` issue template as the intake entry point, then move serious conversations into the private operating tracker.

## 15-Minute Evaluation Path

1. Send the benchmark comparison that matches their problem.
2. Point them to the public install path or Homebrew formula.
3. Ask them to run `profitctl compare` once or share the closest pricing shape.
4. If the result is useful, ask for real inputs so the model can be calibrated.
5. Decide whether they need docs help, a feature, or a design-partner call.

## Minimum Qualification Questions

Ask only what is needed to map their scenario:

- What are you pricing today?
- Is revenue mostly recurring, mostly one-time, or mixed?
- Do free users, paid monthly users, and paid annual users behave differently?
- What is the contract shape you care about most right now?
- What would make this useful enough to install internally?

## Intake Notes To Capture

- company and contact
- current pricing shape
- benchmark scenario they reacted to
- whether they need install help
- whether they want calibration help
- whether they are open to a design-partner call

## Follow-Up Paths

- install help: point them to the release and install docs
- calibration help: ask for the minimum exported inputs
- design-partner pilot: agree on one contract shape to test and one success criterion

## What We Need From The First 5 Evaluations

- which benchmark pair made sense immediately
- whether install worked without intervention
- which metric they trusted first
- which input they could not map into the current config
- whether they would use `compare` before a real pricing or contract decision

## Good Design-Partner Criteria

Treat a user as a good fit if they:

- already have a pricing or packaging problem
- can share real inputs instead of hypotheticals
- want to compare shapes before shipping
- care about recurring economics, not just a vanity simulation

## Do Not Overbuild

Do not create a heavyweight sales process yet.

The only thing we need is a consistent loop from:

benchmark -> install -> compare -> calibrate -> follow-up
