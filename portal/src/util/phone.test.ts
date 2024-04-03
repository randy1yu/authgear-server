/* global it, expect */
import {
  cleanRawInputValue,
  trimCountryCallingCode,
  makePartialValue,
} from "./phone";

it("cleanRawInputValue", () => {
  expect(cleanRawInputValue("1234 1234")).toEqual("12341234");
  expect(cleanRawInputValue("asdf")).toEqual("");
  expect(cleanRawInputValue("我")).toEqual("");
  expect(cleanRawInputValue("+852+852")).toEqual("+852852");
});

it("trimCountryCallingCode", () => {
  expect(trimCountryCallingCode("+", "852")).toEqual("");
  expect(trimCountryCallingCode("+852", "852")).toEqual("");
  expect(trimCountryCallingCode("+85298", "852")).toEqual("98");
  expect(trimCountryCallingCode("98765432", "852")).toEqual("98765432");
});

it("makePartialValue", () => {
  expect(makePartialValue("+", "852")).toEqual("+852");
  expect(makePartialValue("+852", "852")).toEqual("+852");
  expect(makePartialValue("123", "852")).toEqual("+852123");
});
