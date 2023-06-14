import { mudConfig, resolveTableId } from "@latticexyz/world/register";

export default mudConfig({
  systems: {
    CreateMatchSystem: {
      name: "creatematch",
      openAccess: true,
    },
    JoinMatchSystem: {
      name: "joinmatch",
      openAccess: true,
    },
    PlaceCardSystem: {
      name: "placecard",
      openAccess: true,
    },
    MoveCardSystem: {
      name: "movecard",
      openAccess: true,
    },
    AttackSystem:{
      name: "attack",
      openAccess: true,
    },
    EndTurnSystem: {
      name: "endturn",
      openAccess: true,
    },
    RegisterSystem: {
      name: "register",
      openAccess: true,
    },
    // Abilities
    MetorSystem: {
      name: "meteor",
      openAccess: true,
    },
    DrainSwordSystem: {
      name: "drainsword",
      openAccess: true,
    },
    PiercingShotSystem: {
      name: "piercingshot",
      openAccess: true,
    },
    WhirlwindAxeSystem: {
      name: "whirlwindaxe",
      openAccess: true,
    },
    SidestepSystem: {
      name: "sidestep",
      openAccess: true,
    },
    CoverSystem: {
      name: "cover",
      openAccess: true,
    },
  },
  tables: {
      // Users
    User: "bytes32",

    // Matches config
    MapConfig: {
      primaryKeys: {},
      schema: {
        width: "uint32",
        height: "uint32",
        maxPlacedCards: "uint32",
      },
    },
    Match: "bool",
    PlayerOne: "bytes32",
    PlayerTwo: "bytes32",

    // Board
    CurrentTurn: "uint32",
    CurrentPlayer: "bytes32",
    CurrentMana: "uint32",
    PlacedCards: {
      schema: {
        p1Cards: "uint32",
        p2Cards: "uint32",
      },
    },

    // Units
    Card: "bool",
    OwnedBy: "bytes32",
    UsedIn: "bytes32", // relation between match and card
    IsBase: "bytes32",
    UnitType: "CardTypes",
    AbilityType: "AbilityTypes",
    AttackDamage: "uint32",
    MaxHp: "uint32",
    CurrentHp: "uint32",
    MovementSpeed: "uint32",
    ActionReady: "bool",
    Position: {
      dataStruct: false,
      schema: {
        placed: "bool",
        gameKey: "bytes32",
        x: "uint32",
        y: "uint32",
      },
    },
    // Skills
    SidestepInitialPosition: {
      dataStruct: false,
      schema: {
        active: "bool",
        x: "uint32",
        y: "uint32",
      },
    },
    // Key is the game id, value card id
    CoverPosition:{
      dataStruct: false,
      schema: {
        coverOneCard: "bytes32",
        coverOnePlayer: "bytes32",
        coverTwoCard: "bytes32",
        coverTwoPlayer: "bytes32",
      },
  },
  },
  enums: {
    // Base MUST be the last value
    CardTypes: ["VaanStrife", "Felguard", "Sakura", "Freya", "Lyra", "Madmartigan", "Base"],
    AbilityTypes: ["Meteor", "DrainSword", "PiercingShot", "WhirlwindAxe", "Sidestep", "Cover"],
  },
  modules: [
    {
      name: "KeysWithValueModule",
      root: true,
      args: [resolveTableId("Position")],
    },
    {
      name: "KeysWithValueModule",
      root: true,
      args: [resolveTableId("UsedIn")],
    },
    {
      name: "KeysWithValueModule",
      root: true,
      args: [resolveTableId("User")],
    },
  ],
});
