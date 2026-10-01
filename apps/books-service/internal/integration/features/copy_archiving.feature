Feature: Archiving physical copies
  Owners retain history when removing unoccupied copies from their shelves.

  Background:
    Given Alice owns an available lendable physical copy

  Scenario Outline: Occupied copies cannot be archived
    Given Bob's loan is "<state>"
    When Alice archives the copy
    Then the action is rejected without changing the copy or requests

    Examples:
      | state          |
      | accepted       |
      | borrowed       |
      | return_pending |

  Scenario: Archive a copy after a completed loan
    Given Bob's loan is "returned"
    When Alice archives the copy
    Then the action succeeds
    And the copy is archived with borrowing history retained
    And Bob's request is "returned"
    And the copy is absent from public shelves
    When Carol requests the copy
    Then the action is rejected without changing the copy or requests

  Scenario: Archiving declines pending requests and retains history
    Given Bob has a pending request for the copy
    And Carol has a pending request for the copy
    When Alice archives the copy
    Then the action succeeds
    And Bob's request is "declined"
    And Carol's request is "declined"
    And the copy is archived with borrowing history retained
    And the copy is absent from public shelves

  Scenario: Another reader cannot archive the owner's copy
    Given Bob has a pending request for the copy
    When Bob archives the copy
    Then the action is rejected without changing the copy or requests
