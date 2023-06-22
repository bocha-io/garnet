// SPDX-License-Identifier: MIT
pragma solidity >=0.8.0;

// Core
import {System} from "@latticexyz/world/src/System.sol";

// Tables
import {Match} from "../codegen/tables/Match.sol";
import {PlayerOne} from "../codegen/tables/PlayerOne.sol";
import {PlayerTwo} from "../codegen/tables/PlayerTwo.sol";
import {addressToEntityKey} from "../addressToEntityKey.sol";

contract SurrenderSystem is System {
    function surrender(bytes32 key) public {
        bool value = Match.get(key);
        require(value, "match not found");

        bytes32 sender = addressToEntityKey(_msgSender());

        bytes32 playerTwo = PlayerTwo.get(key);
        bytes32 playerOne = PlayerOne.get(key);
        require(
            playerOne == sender || playerTwo == sender, "you are trying to surrender a game that you are not part of it"
        );

        PlayerOne.deleteRecord(key);
        PlayerTwo.deleteRecord(key);
        Match.deleteRecord(key);
    }
}
