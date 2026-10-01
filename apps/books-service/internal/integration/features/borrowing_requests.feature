Feature: Requesting and reserving physical copies
  Readers request a particular owner's copy, rather than a catalog title.

  Background:
    Given Alice owns an available lendable physical copy

  Scenario: Request an available copy
    When Bob requests the copy
    Then the action succeeds
    And Bob's request is "pending"
    And the copy is "available"

  Scenario: Owners cannot borrow their own copy
    When Alice requests the copy
    Then the action is rejected without changing the copy or requests

  Scenario: A borrower cannot create duplicate open requests
    Given Bob has a pending request for the copy
    When Bob requests the copy
    Then the action is rejected without changing the copy or requests

  Scenario: Approval reserves the copy and declines competing requests
    Given Bob has a pending request for the copy
    And Carol has a pending request for the copy
    When Alice accepts Bob's request
    Then the action succeeds
    And Bob's request is "accepted"
    And Carol's request is "declined"
    And the copy is "reserved"

  Scenario Outline: Only the owner approves or declines requests
    Given Bob has a pending request for the copy
    When <actor> <action> Bob's request
    Then the action is rejected without changing the copy or requests

    Examples:
      | actor | action   |
      | Bob   | accepts  |
      | Carol | accepts  |
      | Bob   | declines |
      | Carol | declines |

  Scenario: Owner declines a pending request
    Given Bob has a pending request for the copy
    When Alice declines Bob's request
    Then the action succeeds
    And Bob's request is "declined"
    And the copy is "available"

  Scenario Outline: An occupied copy cannot receive new requests
    Given Bob's loan is "<state>"
    When Carol requests the copy
    Then the action is rejected without changing the copy or requests

    Examples:
      | state          |
      | accepted       |
      | borrowed       |
      | return_pending |
