// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

// Core
import {System} from "@latticexyz/world/src/System.sol";

// Tables
import {Match} from "../codegen/tables/Match.sol";
import {PlayerOne} from "../codegen/tables/PlayerOne.sol";
import {PlayerTwo} from "../codegen/tables/PlayerTwo.sol";
import {CurrentPlayer} from "../codegen/tables/CurrentPlayer.sol";
import {addressToEntityKey} from "../addressToEntityKey.sol";

contract SurrenderSystem is System {
    function surrender(bytes32 key) public {
        bool value = Match.get(key);
        require(value, "match not found");

        bytes32 currentPlayer = CurrentPlayer.get(key);
        require(addressToEntityKey(_msgSender()) == currentPlayer, "current player must be the sender");

        bytes32 playerTwo = PlayerTwo.get(key);
        bytes32 playerOne = PlayerOne.get(key);

        PlayerOne.deleteRecord(key);
        PlayerTwo.deleteRecord(key);
        Match.deleteRecord(key);
    }
}
