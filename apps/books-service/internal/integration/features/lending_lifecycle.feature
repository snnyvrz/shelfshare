Feature: Lending, returning and cancelling a physical copy
  Only owner-confirmed receipt releases a borrowed copy.

  Background:
    Given Alice owns an available lendable physical copy

  Scenario: Owner confirms handover
    Given Bob's loan is "accepted"
    When Alice confirms handover of Bob's request
    Then the action succeeds
    And Bob's request is "borrowed"
    And the copy is "on loan"

  Scenario: Borrower return does not release availability
    Given Bob's loan is "borrowed"
    When Bob initiates return of Bob's request
    Then the action succeeds
    And Bob's request is "return_pending"
    And the copy is "on loan"
    When Carol requests the copy
    Then the action is rejected without changing the copy or requests

  Scenario: Owner receipt completes the loan and permits borrowing again
    Given Bob's loan is "return_pending"
    When Alice confirms receipt of Bob's request
    Then the action succeeds
    And Bob's request is "returned"
    And the copy is "available"
    When Carol requests the copy
    Then the action succeeds
    And Carol's request is "pending"

  Scenario Outline: Only the correct participant performs loan actions
    Given Bob's loan is "<state>"
    When <actor> <action> Bob's request
    Then the action is rejected without changing the copy or requests

    Examples:
      | state          | actor | action              |
      | accepted       | Bob   | confirms handover of |
      | accepted       | Carol | confirms handover of |
      | borrowed       | Alice | initiates return of  |
      | borrowed       | Carol | initiates return of  |
      | return_pending | Bob   | confirms receipt of  |
      | return_pending | Carol | confirms receipt of  |

  Scenario Outline: Actions must follow the lifecycle order
    Given Bob's loan is "<state>"
    When <actor> <action> Bob's request
    Then the action is rejected without changing the copy or requests

    Examples:
      | state    | actor | action               |
      | pending  | Alice | confirms handover of |
      | accepted | Bob   | initiates return of  |
      | borrowed | Alice | confirms receipt of  |
      | returned | Alice | accepts              |
      | returned | Alice | confirms handover of |

  Scenario Outline: Participants cancel before handover
    Given Bob's loan is "<state>"
    When <actor> cancels Bob's request
    Then the action succeeds
    And Bob's request is "cancelled"
    And the copy is "available"

    Examples:
      | state    | actor |
      | pending  | Bob   |
      | accepted | Alice |
      | accepted | Bob   |

  Scenario Outline: Cancellation is restricted to eligible participants and states
    Given Bob's loan is "<state>"
    When <actor> cancels Bob's request
    Then the action is rejected without changing the copy or requests

    Examples:
      | state          | actor |
      | pending        | Alice |
      | accepted       | Carol |
      | borrowed       | Alice |
      | borrowed       | Bob   |
      | return_pending | Alice |
      | return_pending | Bob   |
      | returned       | Bob   |

  Scenario Outline: Authorized retries do not change persisted state
    Given Bob's loan is "<state>"
    When <actor> <action> Bob's request
    Then the action succeeds
    When <actor> <action> Bob's request
    Then the action succeeds without changing the copy or requests

    Examples:
      | state          | actor | action               |
      | pending        | Alice | accepts              |
      | pending        | Alice | declines             |
      | pending        | Bob   | cancels              |
      | accepted       | Alice | confirms handover of |
      | borrowed       | Bob   | initiates return of  |
      | return_pending | Alice | confirms receipt of  |
