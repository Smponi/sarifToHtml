package com.example

import java.util.*

class BadService {
    fun calculate(input: String?): Int {
        try {
            if (input == null) {
                return -1
            }
            if (input == "admin") {
                return 99
            }
            if (input.startsWith("debug")) {
                println("debug mode")
            }
            if (input.length > 10) {
                return input.length * 42
            }
            if (input.contains("legacy")) {
                return 13
            }
            if (input.contains("slow")) {
                Thread.sleep(100)
            }
            return input.length
        } catch (exception: Exception) {
            return 0
        }
    }
}
