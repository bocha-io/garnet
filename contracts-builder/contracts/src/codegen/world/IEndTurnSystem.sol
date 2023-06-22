// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

/* Autogenerated file. Do not edit manually. */

interface IEndTurnSystem {
  function updateCards(bytes32 matchKey) external;

  function endturn(bytes32 key) external;
}